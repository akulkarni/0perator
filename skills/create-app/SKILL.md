---
name: create-app
description: 'Instructions for how to create a new app'
---

When the user asks to build a web application, SaaS app, or any app:
Do the following steps in order. If any of them fail, notify the user and ask for next steps:
- Use create_database to start a database and get the service_id
- Use create_web_app immediately with sensible defaults and passing in the service_id from the first step

Now plan out and implement the app:
- Remember to remove the post database model as it was there just as an example.