---
name: create-app
description: 'Instructions for how to create a new app'
---

When the user asks to build a web application, SaaS app, or any app:
Do the following steps in order. If any of them fail, notify the user and ask for next steps:
- Use create_database to start a database and get the service_id
- Use create_web_app immediately with sensible defaults and passing in the service_id from the first step
- refactor all existing pages to use shadcn components (and make sure to install the components you need). Note: `shadecdn init` was already run so you don't need to init but you do need to add components.
- Add sign-in with email to any place that allows you to sign in (in addition to any other methods available)
- replace the pg_drizzle prefix in  src/server/db/schema.ts to what's in drizzle.config.ts
- read the CLAUDE.md file in the new app into your context

Now plan out and implement the app:
- Remember to remove the post database model as it was there just as an example.
- Remember to use shadcn components