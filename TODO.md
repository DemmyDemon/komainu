
# Commands
| Done | Command  | Arguments       | Group   | Function                                                               |
| ---- | -------- | --------------- | ------- | ---------------------------------------------------------------------- |
|      | grant    | string role     | admin   | Grant access to command group `string` to role `role`                  |
|      | revoke   | string role     | admin   | Revoke access to command group `string` to role `role`                 |
|      | access   |                 | admin   | List the access grants (and automagically purge invalid ones)          |
|      | faq      | string          | faquser | Look up the value for key `string` in the FAQ                          |
|      | faqon    | string1 string2 | faq     | Add or replace value `string2` for the key `string1` in the FAQ        |
|      | faqoff   | string          | faq     | Delete the key `string` from the FAQ                                   |
|      | faqlist  |                 | faq     | LIst the FAQ topics                                                    |
|      | vote     | float string    | vote    | Start a vote on the question `string` that will run for `float` hours. |
|      | seen     | member          | seen    | Look up when `member` was last seen saying anything.                   |
|      | inactive | int             | seen    | Look up who has not said anthing for `int` days.                       |

# Storage
| Done | Stores what   |
| :--: | ------------- |
|  X   | Configuration |
|      | Access rights |
|      | FAQ items     |
|      | Votes         |
|  X   | Seen times    |

# Voting

I can use the emoji 👎 and 👍 for positive and negative votes.