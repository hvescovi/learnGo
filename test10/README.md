# test9
## services, in PODs, with slaves

Inspired by the last phrases of this link:
https://github.com/kubernetes/kubernetes/tree/master/examples/guestbook#step-two-fire-up-the-redis-master-service

A leader is a POD running under a service leader, svcleader.

Slaves will run in PODs running under a service slave, called svcslave.

When a slave starts, it will "connect" in the svcleader and will "hear" for updates. Like shared registers, maybe.




===

*old*

When a server receive a request, it executes it and look for another replicas. If another replicas were found, the server asks these servers to also run the request.

Each time a server receives a requests, it searches for another replicas. At that moment the view of the group can change, on the view point of a server.

**All** replicas are required to answer: the server who contacts the others replicas will wait for that.

*Maybe it is possible the application (server) asks kubernetes how many replicas of the application are running. This could make easier the discovery of the group.
