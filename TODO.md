
# Commands
| Done | Command       | Arguments       | Group    | Function                                                               |
| :--: | ------------- | --------------- | -------- | ---------------------------------------------------------------------- |
|  X   | access grant  | string role     | access   | Grant access to command group `string` to role `role`                  |
|  X   | access revoke | string role     | access   | Revoke access to command group `string` to role `role`                 |
|  X   | access list   |                 | access   | List the access grants (and automagically purge invalid ones)          |
|  X   | faq           | string          | faquser  | Look up the value for key `string` in the FAQ                          |
|  X   | faqset add    | string1 string2 | faqadmin | Add or replace value `string2` for the key `string1` in the FAQ        |
|  X   | faqset remove | string          | faqadmin | Delete the key `string` from the FAQ                                   |
|  X   | faqset list   |                 | faqadmin | LIst the FAQ topics                                                    |
|      | vote          | float string    | vote     | Start a vote on the question `string` that will run for `float` hours. |
|  X   | seen          | member          | seen     | Look up when `member` was last seen saying anything.                   |
|  X   | inactive      | int             | seen     | Look up who has not said anthing for `int` days.                       |

# Storage
| Done | Stores what   |
| :--: | ------------- |
|  X   | Configuration |
|  X   | Access rights |
|  X   | FAQ items     |
|      | Votes         |
|  X   | Seen times    |
|      | Throttling    |

# Voting
I can use the emoji 👎 and 👍 for positive and negative votes.

# Throttling
To avoid mass spamming of things like `/faq`, there might be need of a throttling function.

Options:
- Per-user ignore commands after processing one.
- Globally ignore commands after processing one.
- Per-user uses-per-minute.
- Global uses-per-minute.
- Different settings for different commands.

The downside to throttling is that if it is in use for a help system, it will limit how often helpers can actually help. Perhaps some sort of excemption system needs to exist?
