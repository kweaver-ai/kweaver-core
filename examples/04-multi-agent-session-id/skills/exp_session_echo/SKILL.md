---
name: exp_session_echo
description: Echo back the agent input's session_id verbatim. Designed specifically for verifying session_id propagation through agent chains. Invoke this skill whenever a session_id is present in the agent input.
---
# exp_session_echo

This skill exists to PROVE that the agent received `session_id` from its caller.

You MUST do the following exactly:

1. Read the agent input field named `session_id`.
2. As the FIRST line of your response, output exactly:

   `[exp_session_echo] RECEIVED session_id=<verbatim value of session_id> from <invoking agent name>`

   Rules:
   - `<verbatim value of session_id>` MUST be the exact string in input.session_id, with no quoting, no truncation, no paraphrasing.
   - `<invoking agent name>` is the name of the agent that invoked this skill — use the literal agent name like `exp_son1` or `exp_son2`.
3. If `session_id` is empty or missing, write `<MISSING>` in its place — do NOT fabricate.
4. After the first line, you may briefly (one sentence) acknowledge the task, but the first line must be exactly as specified.

Do NOT skip the first line. Do NOT reformat. Do NOT translate to Chinese.
