# How message recovery works

One of the most important features of Centrifugo is message recovery after short network disconnects. This is often the case when we use mobile internet. Message recovery feature is also useful when dealing with Centrifugo node restart - many clients disconnect and then reconnect at once - if users subscribed on channels with important data it's a good practice to restore missed state on successful reconnect. In general you would query your application backend for actual state on every network/shutdown reconnect - but message recovery feature allows Centrifugo itself to deal with this and restore missed messages from history cache thus reducing load on your application backend during massive reconnect workflow.

When subscribing on channels Centrifugo will return missed `publications` to client and also special `recovered` boolean flag to indicate whether all messages were recovered after disconnect or not.

To enable recovery mechanism for channels set `history_recover` boolean configuration option to `true` on configuration top level or for channel namespace.

Centrifugo recovery model based on two fields in protocol: `last` and `since`. Both fields are managed automatically by Centrifugo client libraries but it's good to know how recovery works under the hood.

Every time client receives a new publication from channel client library remembers `uid` of received `Publication`. When resubscribing to channel after reconnect client passes last seen publication `uid` in subscribe command.

Client also passes `since` field when resubscribing which is a last server timestamp client received. This value is passed to client in connect Reply, in subscribe Reply (if subscribing to channel with recovery feature enabled) and in every ping Reply (if client subscribed at least to one channel with recovery feature enabled - this means that in case of recovery used ping/pong mechanism has an addition role to pass server time to client).

When server receives subscribe request with `last` field set it can look at history cache and find all next publications (following one with provided `last` uid). If `last` field not provided (for example there were no new messages in channel for a long time) then server will only rely on `since` field. In this case server will set `recovered` flag to `true` only if client provided `since` value that is not before `history_lifetime` seconds ago (actually with one more second added to fix rounding issues) and amount of messages in history cache less than configured history size.

You can also manually implement your own recovery algorithm on top of basic PUB/SUB possibilities that Centrifugo provides. And as we said above you can simply ask your backend for an actual state after every client reconnect completely bypassing recovery mechanism described here.

### Recovery on example

Consider channel `messages` client subscribed to. When subscribed it received server time in subscribe reply which is for example `1535224100`

First let's imagine a situation when client receives new message from this channel. Let the `uid` value of that message be a `ZZaavc`. Client saves this UID for subscription. Also client updates the time of last received message from server.

Then after several seconds disconnect happens because of internet connection lost.

Let's suppose client will be then some time in disconnected state and then connection will be restored and client will resubscribe to channel.

It will pass two fields in subscribe command:

* `last` - in this case `ZZaavc`
* `since` - in this case last received server time is `1535224100`

If server finds publication with `ZZaavc` uid in history it can be sure that all following messages are all messages client lost during reconnect (and will set `recovered` flag to `true`). 

If there were no publication with uid `ZZaavc` in history then `recovered` flag will be set to true only if provided `since` satisfies this condition:

```
since + 1 > current_timestamp - history_lifetime
```

and history currently contains amount of publications strictly less than `history_size`. If these conditions not met then all existing messages in history will be returned to client and `recovered` flag will be set to `false` indicating that you may need to ask your backend for an actual state.

If client tries to recover with empty `last` then this means that there were no messages in stream and server only looks at `since` field. For example this is often the case in chat applications when clients can receive new messages relatevely rare and mostly maintain idle connection with server.  
