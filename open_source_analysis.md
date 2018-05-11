A list of popular open-source Go programs, analyzing how they use concurrency:

### Go's own HTTP package

Go's built-in HTTP package utilizes Go's concurrency.

Uses of concurrency in Go's HTTP package:

* To serve multiple incoming requests
* To run any routines that must be run before the server is shut down
* To listen on a network without blocking


### Gin

Gin is a web framework that advertises itself on its performance.
I was surprised to find that this increase in performance had little to do with utilization of concurrency.
Instead, the increase was due to the use of another Go package, [httprouter](https://github.com/julienschmidt/httprouter).
httprouter itself does not utilize Go's concurrency model.
Gin does not utilize much concurrency in its own code because it uses Go's built-in HTTP pakage.

Uses of concurrency in Gin:

* Run multiple web services at the same time

### Syncthing

A program for continuously synchronizing files between multiple machines.

Uses of concurrency in Syncthing:
* To asynchronously update a database
* To separate tasks through use of channels, e.g. separating scanning a list of paths from collecting the path names
* To limit the amount of a certain task that can run concurrently by using channels
* To run a large number of the same task in parallel
* To schedule setup events
* To send packets without blocking the main thread
* To receive packets without blocking the main thread
* To detect deadlocks

