---
name: create-app
description: 'Instructions for how to create a new app'
---

When the user asks to build a web application, SaaS app, or any app:
Do the following steps in order. If any of them fail, notify the user and ask for next steps:
- Use create_database to start a database and get the service_id
- Figure out if this is a multi-user app. If it's not clear from the prompt ask the user.
- If it is a multi-user app ask what authentication methods the user wants. Ask user to pick from one or more: email signup, github auth provider, google auth provider. 
- Use create_web_app immediately with sensible defaults and passing in the service_id from the create_database call and enabling auth if it's a multi-user app.
- refactor all existing pages to use shadcn components (and make sure to install the components you need). Note: `shadecdn init` was already run so you don't need to init but you do need to add components.
- verify that all pages use the shadcn colors and not what comes with the t3 template. For example, change `bg-gradient-to-b from-slate-900 to-slate-800` to `bg-background`. Do that for every page.
- if using auth, change the sign-in form to support all the methods the user requested. Make this a component. Also edit src/server/better-auth/config.ts to reflect the user's choice.
- replace the pg_drizzle prefix in  src/server/db/schema.ts to what's in drizzle.config.ts
- read the CLAUDE.md file in the new app into your context

Now plan out and implement the app:
- Remember to remove the post database model as it was there just as an example.
- Remember to use shadcn components