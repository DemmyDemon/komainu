
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
|  X   | Throttling    |

# Voting
After discussing votes with the people that will be using the bot, the idea of secret votes has come up.

Buttons with ephemeral responses to clicking them should be able to record the vote without showing who voted what.

# Throttling
The first iteration is implemented.

It has "tokens" that get spent for user and channel *separately*, and they are "refunded" at a 10 second interval. Probably needs more work, but I think it has to be in actual use first.

Considering this TODONE for now.

