# Komainu 101

Nearly all the interaction done with Komainu uses the so-called "slash commands". This is entirely unrelated to horror films, and refers to the commands starting with a `/`, a forward slash.

## Commands

The commands are:

### /activerole

This allows you to set a role that is given to those that speak, and is then taken away when they haven't spoken for a while. It takes two arguments: `role` and `days`.

The `role` is any `@role` defined in your Discord guild. `days` is a floating point number. Setting `days` to zero disables the function entirely.

Example:  `/activerole @Chatterbox 30`  
Gives anyone that speaks the Chatterbox role for 30 days.

"Speaks" refers to regular text chat only. It does not count status changes or reactions to messages, only to sending messages of your own. Note that this only counts messages the bot has seen, so any message in a channel the bot doesn't have access to doesn't count. If the bot was offline when the message was sent it is not counted either.

### /ateball

This is just for fun. It's like a magic 8-ball, but food themed, for some weird reason. It takes a single argument: `question`.

The question is required, as it is echoed back when answering the question, but it does not impact the answer in any way what so ever. A random answer will be picked, along with a random food-related emoji.

Example: `/ateball Will my crush finally notice me?`  
This will make the bot crush your dreams, possibly with a food-related pun.

### /deletelog

This allows you to have the bot monitor for messages being deleted, and put a notice about it (possibly containing the message) in the channel of your choice. It takes a single argument:  `channel`.

The `channel` is any already existing channel that the bot has access to sending messages in. If you leave this blank, the feature is turned off.

Example:  `/deletelog #deleted-log`  
All deleted messages will now be logged in the `#deleted-log` channel.

Note that the bot does not actually keep a record of all messages it sees. This would be a huge invasion of privacy. Instead, it keeps messages it sees in memory ("cache") for a while. The length of that while depends entirely on how much activety there is, but it could be several weeks. Any time the bot is restarted, the messages are entirely lost immediately, and no attempt is made to retrieve them from Discord.

In the event of a message being deleted that is *not* still in the cache, it will simply log that an "unknown message" was deleted and where it was deleted from, with no further details available.

### /faq

This allows you to look up a previously stored FAQ topic. May be handy for that question that is asked very frequently, like a list of what channels do what, or simply as a "fun fact"-regurgitator regardless of how frequently the question is actually asked. It takes a single argument:  `topic`.

The `topic` is a keyword, or phrase, that was specified when the topic was saved.

Example: `/faq horseradish`  
This will look up the topic `horseradish` and display the text associated with it, if any.

The bot will make some effort to help you by attempting auto-complete your topic.

### /faqset

This one is a bit complicated, as it is divided into sub-commands.

#### /faqset add

This allows you to add a topic to the list of FAQ topics. It takes a single argument:  `topic`.

After entering this command, you will be presented with a modal dialog to enter the details of the topic. If the topic already exists, the existing text will be displayed for you to edit.

Example:  `/faqset add horseradish`  
This will present you with a text box where you can describe what a horseradish is and why it's relevant.

#### /faqset remove

This allows you to remove a topic from the list of FAQ topics. It takes a single argument: `topic`.

Example: `/faqset remove horseradish`  
This will for ever erase your witty and insightful essay on horseradishes and their many uses in gaming culture.

### /faqset list

This allows you to list all the FAQ topics. It takes no arguments.

Example: `/faqset list`  
This will list all the topics known to the bot at this moment.

### /inactive

This allows you to check who has been inactive in your Discord guild. The bot jots down the time when someone sends a message, and compares that to the current time when asked. The result is text file it presents for you to view. It takes a single argument: `days`.

In this context `days` is an integer number of 24 hour periods from the current second.

Example: `/inactive 30`  
This will present you with a text file named `inactive_report_(current date here).txt`, containing everyone that has not sent any messages in the past 30 days, including those that have never sent any messages. Where appicable it will tell you how long they have been inactive, in whole days.

Note that this only counts messages the bot has seen, so any message in a channel the bot doesn't have access to doesn't count. If the bot was offline when the message was sent it is not counted either.

### /neverseen

This is very similar to `/inactive`, but lists only those that have never been seen. It does not accept any arguments.

Example:  `/neverseen`  
This will present you with a text file named `never_seen_report_(current date here).txt`, containing everyone currently in the Discord guild that the bot has not yet seen send any messages. Alongside the user will be their join date so you know if they've been lurking for 6 months or 3 minutes.

### /seeeveryone

This will actively subvert `/inactive` and `/neverseen` and store every last current member of the Discord guild as if they've sent a message *right now*. It takes no arguments.

Example:  `/seeeveryone`  
Eeeeeveryone are now counted as active!

Note that this does *not* grant the `/activerole` if one is set, it only counts towards `/inactive` and `/neverseen`, with one exception:  If they already have the active role, their countdown to losing it will start *now*.

### /seen

Much like `/inactive` and `/neverseen`, this will check when someone last sent a message, but the lookup is specific to a single person. It takes one argument: `user`.

Example: `/seen @Demonen`  
This will tell you when `@Demonen` last sent a message in this Discord guild.

### /vote

This is for initating votes. It will *not* disclose who voted what. It takes a single artument:  `length`.

In this context, `length` is the vote length in *days*, as a *floating point* number of 24 hour periods.

Example: `/vote 0.5`  
This will initiate a vote that will run for 12 hours before closing.

You will be prompted for a text to describe what is being voted on, and for a list of options. The options list is just a large input field, where each line is a separate option.  
The options can be up to 100 characters long. Anything longer than that will be cut off without warning.  
There can be a maximum of 25 options. Any more will also be cut off without warning.
