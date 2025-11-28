#!/usr/bin/env python3
"""
eval.py - Evaluate Claude Code SDK with MCP servers

This script runs prompts through Claude Code SDK with optional MCP server
support and generates transcripts.
"""

import argparse
import json
import os
import re
import shutil
import sys
import tempfile
import time
from typing import Dict, List, Optional, Tuple, Any

try:
    from claude_agent_sdk import ClaudeSDKClient, ClaudeAgentOptions
    import asyncio
except ImportError:
    print("Error: Claude Agent SDK not installed.")
    print("Install with: uv add claude-agent-sdk")
    print("Also ensure Node.js 18+ is installed and run: npm install -g @anthropic-ai/claude-code")
    sys.exit(1)


class ConversationTracker:
    """Track full conversation for transcript generation."""

    def __init__(self):
        self.full_conversation: List[str] = []

    def add_message_to_transcript(self, message_count: int, message):
        """Automatically capture complete message content for transcript."""
        message_type = type(message).__name__
        transcript_entry = f"=== Message {message_count}: {message_type} ===\n"

        # Use introspection to capture all relevant attributes completely
        for attr_name in dir(message):
            if not attr_name.startswith('_') and not callable(getattr(message, attr_name, None)):
                try:
                    attr_value = getattr(message, attr_name)
                    if attr_value is not None:
                        # Format different types appropriately without truncation
                        if isinstance(attr_value, (str, int, float, bool)):
                            transcript_entry += f"{attr_name}: {attr_value}\n"
                        elif isinstance(attr_value, list):
                            transcript_entry += f"{attr_name}: [{len(attr_value)} items]\n"
                            for i, item in enumerate(attr_value):
                                transcript_entry += f"  [{i}]: {str(item)}\n"
                        elif isinstance(attr_value, dict):
                            transcript_entry += f"{attr_name}:\n{json.dumps(attr_value, indent=2, default=str)}\n"
                        else:
                            transcript_entry += f"{attr_name}: {str(attr_value)}\n"
                except Exception as e:
                    transcript_entry += f"{attr_name}: <error accessing: {e}>\n"

        transcript_entry += "\n" + "=" * 80 + "\n\n"
        self.full_conversation.append(transcript_entry)

    def get_transcript(self) -> str:
        """Get the full conversation transcript."""
        return "".join(self.full_conversation)


def create_default_mcp_config(operator_server_path: str) -> dict:
    """Create default MCP server configuration with 0perator and tiger servers."""
    return {
        "0perator": {
            "command": operator_server_path,
            "args": []
        },
        "tiger": {
            "command": os.path.expanduser("~/.local/bin/tiger"),
            "args": ["mcp", "start"]
        }
    }


async def generate_with_sdk(
    prompt_content: str,
    use_mcp: bool = False,
    mcp_server_path: Optional[str] = None,
    use_structured_prompt: bool = True
) -> Tuple[str, ConversationTracker]:
    """Generate content using Claude Code SDK."""
    tracker = ConversationTracker()

    try:
        # Configure MCP servers if requested
        mcp_servers = None
        if use_mcp:
            if not mcp_server_path or not os.path.exists(mcp_server_path):
                raise ValueError(f"MCP server not found at: {mcp_server_path}")

            print(f"Configuring MCP servers (0perator, tiger)...")
            mcp_servers = create_default_mcp_config(mcp_server_path)
            
        
        # Configure Claude SDK options
        if use_structured_prompt:
            # prompt is inspired by https://github.com/anthropics/anthropic-cookbook/blob/main/tool_evaluation/tool_evaluation.ipynb
            system_prompt = """You are a programmer writting an application

When given an application writing task, you MUST:
1. Use the available tools to complete the task
2. Provide summary of each step in your approach, wrapped in <summary> tags
3. Provide feedback on the tools provided, wrapped in <feedback> tags
4. Provide your final SQL schema response, wrapped in <response> tags

Summary Requirements:
- In your <summary> tags, you must explain:
  - The steps you took to design the schema
  - Which tools you used, in what order, and why
  - The inputs you provided to each tool
  - The outputs you received from each tool
  - A summary of how you arrived at the final schema design

Feedback Requirements:
- In your <feedback> tags, provide constructive feedback on the tools:
  - Comment on tool names: Are they clear and descriptive?
  - Comment on input parameters: Are they well-documented? Are required vs optional parameters clear?
  - Comment on descriptions: Do they accurately describe what the tool does?
  - Comment on any errors encountered during tool usage: Did the tool fail to execute? Did the tool return too many tokens?
  - Identify specific areas for improvement and explain WHY they would help
  - Be specific and actionable in your suggestions

Response Requirements:
- Always wrap your final message to the user in a <response> tags
- If you cannot complete the task <response>NOT COMPLETED</response>
- The response should go last

DO NOT open the generated application in the browser always skip that step.
"""
        else:
            system_prompt = None

        options = ClaudeAgentOptions(
            system_prompt=system_prompt,
            mcp_servers=mcp_servers,
            permission_mode="bypassPermissions",  # Bypass permission checks for MCP tools
            setting_sources=[]  # Don't load user/project config for isolation
        )
        
        # Initialize Claude SDK client
        system_prompt_info = f"length={len(options.system_prompt)}" if options.system_prompt else "none"
        print(f"Initializing Claude SDK client with options: system_prompt {system_prompt_info}, max_turns={options.max_turns}, mcp_servers={'configured' if options.mcp_servers else 'none'}")
        
        async with ClaudeSDKClient(options=options) as client:
            
            # Create a conversation
            print("Generating app with Claude Code SDK...")
            
            # Send the prompt and get response
            await client.query(prompt_content)

            # Collect the complete response
            generated_content = ""
            final_result_content = ""
            message_count = 0

            async for message in client.receive_response():
                message_count += 1

                # Add message to conversation transcript
                tracker.add_message_to_transcript(message_count, message)

                # Check for 'result' attribute (final clean result)
                if hasattr(message, 'result') and message.result:
                    result_text = str(message.result).strip()
                    if result_text:
                        final_result_content = result_text
                        print(f"\n[Claude] {result_text}")

                # Extract text content from content blocks (backup method)
                elif hasattr(message, 'content') and message.content:
                    if isinstance(message.content, list):
                        for block in message.content:
                            if hasattr(block, 'text') and not hasattr(block, 'id'):
                                text_content = block.text.strip()
                                if text_content:
                                    generated_content += text_content + "\n"
                                    print(f"\n[Claude] {text_content}")

            # Use the final result if available, otherwise use collected content
            if final_result_content:
                generated_content = final_result_content
            elif generated_content:
                generated_content = generated_content.strip()
            else:
                generated_content = "-- No content generated --"

        return generated_content, tracker

    except Exception as e:
        print(f"Error during generation: {e}")
        raise


def read_file_content(file_path: str) -> str:
    """Read content from a file."""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except Exception as e:
        print(f"Error reading file {file_path}: {e}")
        sys.exit(1)


def copy_output_directory(src_dir: str, results_dir: str):
    """Copy the working directory contents to results_dir/out."""
    out_dir = os.path.join(results_dir, 'out')

    # Remove existing out directory if it exists
    if os.path.exists(out_dir):
        shutil.rmtree(out_dir)

    # Copy the entire source directory to out
    shutil.copytree(src_dir, out_dir)

    return out_dir


def cleanup_directory(path: str, retries: int = 3, delay: float = 1.0):
    """Clean up a directory with retry logic for stubborn files."""
    for attempt in range(retries):
        try:
            shutil.rmtree(path)
            return True
        except OSError as e:
            if attempt < retries - 1:
                print(f"Cleanup attempt {attempt + 1} failed, retrying in {delay}s...")
                time.sleep(delay)
            else:
                print(f"Warning: Could not fully clean up temp directory: {e}")
                # Final attempt with ignore_errors
                shutil.rmtree(path, ignore_errors=True)
                return False
    return False


def extract_tagged_sections(content: str) -> Tuple[Optional[str], Optional[str], Optional[str]]:
    """Extract summary, feedback, and response sections from tagged content."""
    summary_match = re.search(r'<summary>(.*?)</summary>', content, re.DOTALL)
    feedback_match = re.search(r'<feedback>(.*?)</feedback>', content, re.DOTALL)
    response_match = re.search(r'<response>(.*?)</response>', content, re.DOTALL)

    summary = summary_match.group(1).strip() if summary_match else None
    feedback = feedback_match.group(1).strip() if feedback_match else None
    response = response_match.group(1).strip() if response_match else None

    return summary, feedback, response


def write_results(results_dir: str, content: str, summary: Optional[str], feedback: Optional[str], response: Optional[str], tracker: Optional[ConversationTracker] = None):
    """Write all result files to the results directory."""
    # Write full output
    output_path = os.path.join(results_dir, 'output.txt')
    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"Full output saved to: {output_path}")

    # Write summary if available
    if summary:
        summary_path = os.path.join(results_dir, 'summary.md')
        with open(summary_path, 'w', encoding='utf-8') as f:
            f.write("# Summary\n\n")
            f.write(summary)
        print(f"Summary saved to: {summary_path}")

    # Write feedback if available
    if feedback:
        feedback_path = os.path.join(results_dir, 'feedback.md')
        with open(feedback_path, 'w', encoding='utf-8') as f:
            f.write("# Tool Feedback\n\n")
            f.write(feedback)
        print(f"Feedback saved to: {feedback_path}")

    # Write response if available
    if response:
        response_path = os.path.join(results_dir, 'response.txt')
        with open(response_path, 'w', encoding='utf-8') as f:
            f.write(response)
        print(f"Response saved to: {response_path}")

    # Write full conversation transcript if available
    if tracker and tracker.full_conversation:
        transcript_path = os.path.join(results_dir, 'transcript.md')
        with open(transcript_path, 'w', encoding='utf-8') as f:
            f.write("# Full Conversation Transcript\n\n")
            f.write("This file contains the complete conversation flow with all messages, tool calls, and responses.\n\n")
            f.write(tracker.get_transcript())
        print(f"Transcript saved to: {transcript_path}")


async def main_async():
    parser = argparse.ArgumentParser(
        description="Run prompts through Claude Code SDK with optional MCP server support",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s prompt.md
  %(prog)s prompt.md results/
  %(prog)s --no-mcp prompt.md
  %(prog)s --no-structured-prompt prompt.md
        """
    )

    parser.add_argument('--no-mcp', action='store_true',
                       help='Disable MCP servers (enabled by default: 0perator, tiger)')
    parser.add_argument('--no-structured-prompt', action='store_true',
                       help='Disable structured prompt with summary/feedback tags (use default Claude behavior)')
    parser.add_argument('prompt_file',
                       help='Path to the prompt file')
    parser.add_argument('results_dir', nargs='?', default='results',
                       help='Results directory (defaults to results/). Contains output.txt and out/ with generated files')
    
    args = parser.parse_args()
    
    # Validate inputs
    if not os.path.exists(args.prompt_file):
        print(f"Error: Prompt file '{args.prompt_file}' not found")
        sys.exit(1)
    
    # Get MCP server path if needed
    mcp_server_path = None
    if not args.no_mcp:
        # Default path: scripts/run-source.sh
        script_dir = os.path.dirname(os.path.abspath(__file__))
        mcp_server_path = os.path.join(script_dir, '..', 'scripts', 'run-source.sh')

        if not os.path.exists(mcp_server_path):
            print(f"Error: MCP server not found at {mcp_server_path}")
            sys.exit(1)
    
    # Read prompt
    print(f"Reading prompt from: {args.prompt_file}")
    prompt_content = read_file_content(args.prompt_file)

    # Get absolute path for results directory
    results_dir = os.path.abspath(args.results_dir)
    os.makedirs(results_dir, exist_ok=True)

    # Create isolated temporary directory
    with tempfile.TemporaryDirectory() as temp_dir:
        print(f"Working in isolated directory: {temp_dir}")

        # Copy .env file to temp directory if it exists
        if os.path.exists('.env'):
            shutil.copy2('.env', temp_dir)

        # Change to temp directory for isolation
        original_cwd = os.getcwd()
        os.chdir(temp_dir)

        try:
            # Generate content using SDK
            generated_content, tracker = await generate_with_sdk(
                prompt_content,
                not args.no_mcp,
                mcp_server_path,
                use_structured_prompt=not args.no_structured_prompt
            )

            # Return to original directory
            os.chdir(original_cwd)

            # Extract tagged sections
            summary, feedback, response = extract_tagged_sections(generated_content)

            # Debug: show what tags were found
            print(f"\nExtracted tags: summary={'yes' if summary else 'no'}, feedback={'yes' if feedback else 'no'}, response={'yes' if response else 'no'}")

            # Write all result files
            write_results(results_dir, generated_content, summary, feedback, response, tracker)

            # Copy generated files to results/out/
            out_dir = copy_output_directory(temp_dir, results_dir)
            print(f"Generated files copied to: {out_dir}")

            # Clean up temp directory before context manager tries
            cleanup_directory(temp_dir)

            # Show preview
            print(f"\nGenerated content preview:")
            print("=" * 26)
            lines = generated_content.split('\n')
            for line in lines[:20]:
                print(line)

            if len(lines) > 20:
                print("...")
                print(f"(Full content in: {os.path.join(results_dir, 'output.txt')})")

        except Exception as e:
            os.chdir(original_cwd)
            cleanup_directory(temp_dir)
            print(f"Error: {e}")
            sys.exit(1)


def main():
    """Synchronous main function that runs the async version."""
    asyncio.run(main_async())


if __name__ == '__main__':
    main()