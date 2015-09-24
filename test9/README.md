# test9
## client and server - state synchronization

When a server receive a request, it executes it and look for another replicas. If another replicas were found, the server asks these servers to also run the request.

Each time a server receives a requests, it searches for another replicas. At that moment the view of the group can change, on the view point of a server.

**All** replicas are required to answer: the server who contacts the others replicas will wait for that.

*Maybe it is possible the application (server) asks kubernetes how many replicas of the application are running. This could make easier the discovery of the group.
