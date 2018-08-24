# How message recovery works

One of the most important features of Centrifugo is message recovery after short network disconnects. This is often the case when we use mobile internet - for example subway tunnel can result in temporary internet connection lost. Message recovery feature is especially useful when dealing with Centrifugo node restart - many clients disconnect and then reconnect at once - if users subscribed on channels with important data it's a good practice to restore missed state on successful reconnect. In general you would query your application backend for actual state on every network/shutdown reconnect - but message recovery feature allows Centrifugo itself to deal with this and restore missed messages from history cache thus reducing load on your application backend during massive reconnect workflow.

When subscribing on channels Centrifugo will return missed `publications` to client and also special `recovered` boolean flag to indicate whether all messages were recovered after disconnect or not.

To enable recovery mechanism for channels set `history_recover` boolean configuration option to `true` on configuration top level or for channel namespace.

Centrifugo recovery model based on two fields in protocol: `last` and `away`. Both fields are managed automatically by Centrifugo client libraries but it's good to know how recovery works under the hood.

Every time client receives a new publication from channel client library remembers `uid` of received `Publication`. When resubscribing to channel after reconnect client passes last seen publication `uid` in subscribe command.

Client also passes `away` field when resubscribing which is a number of seconds (rounded to the upper value) client has not received any messages from server. From this perspective periodic PING/PONG frames have a very important role as PONG replies allow to monotonically update the time of last message from server received by client. Client also adds timeout values in seconds to this interval to compensate possible latencies when communicating with server.

When server receives subscribe request with `last` field set it can look at history cache and find all next publications (following one with provided `last` uid). If `last` field not provided (for example there were no new messages in channel for a long time) then server will only rely on `away` field. In this case server will set `recovered` flag to `true` only if client was away for an interval of time that is not larger than history lifetime and amount of messages in history cache less than configured history size.

Message recovery in Centrifugo is not a 100% bulletproof scheme (as it has some assumptions regarding to time) but it should work for significant amount of real life cases without message loss. If you need more reliability in channel message recovery then you can manually implement your own recovery algorithm on top of basic PUB/SUB possibilities that Centrifugo provides. You can always simply ask your backend for an actual state after client reconnect completely bypassing recovery mechanism described here.

### Recovery on example

Consider channel `messages` client subscribed to.

First let's imagine a situation when client receives new message from this channel. Let the `uid` value of that message be a `ZZaavc`. Client saves this UID for subscription. Also client updates the time of last received message from server.

Then after `4` seconds disconnect happens because of internet connection lost.

Let's suppose client will be then `6` seconds in disconnected state and then connection will be restored and client will resubscribe on channel.

It will pass two fields in subscribe command:

* `last` - in this case `ZZaavc`
* `away` - in this case `4` + `6` + client request timeout. I.e `15` in case of client request timeout of `5 seconds`.

If server finds publication with `ZZaavc` uid in history it can be sure that all following messages are all messages client lost during reconnect (and will set `recovered` flag to `true`). 

If there were no publication with uid `ZZaavc` in history then `recovered` flag will be set to true only if provided `away` value is less than channel `history_lifetime` and history currently contains amount of publications strictly less than `history_size`. Actually server adds one more second to `away` value to compensate rounding errors. This means that if `history_size` is `10` and `history_lifetime` is `60` then client must reconnect in up to `44` seconds and there must be no more than `9` messages in history to be sure all missed publications were successfully recovered. If this condition not met then all existing messages in history will be returned to client and `recovered` flag will be set to `false` indicating that you may need to ask your backend for an actual state.
