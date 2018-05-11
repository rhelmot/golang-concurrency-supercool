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

### CockroachDB

A scalable SQL database.

Uses of concurrency in Cockroach.db:

* In test, to simulate multiple concurrent edits
* To initiate logging without blocking the main thread
* To run tasks asynchronously and limit the number of tasks through a semaphore channel
* To set a timeout for a function
* To wait for a function to return and run cleanup for it without blocking the main thread
* To serve multiple network connections concurrently
* To run one of a certain function for every processor

### Hugo

A framework for rendering static websites.

Uses of concurrency in Hugo:

* To allow an object's run function to run concurrently with the main thread
* To render multiple pages concurrently
* To separate tasks through use of channels, e.g. separating error collecting from the function that produces the errors
* Run multiple web services at the same time
* To initiate logging without blocking the main thread
* To set a timeout for a function

## Primitive

A program that reproduces images with geometric primitives.
Utilizes very little concurrency.

Uses of concurrency in Primitive:

* To run multiple workers concurrently

### GoTTY

Allows you to run terminal commands as web applications.

Uses of concurrency in GoTTY:

* To wait for error signals from a running server
* To wait for a function to return and run cleanup for it without blocking the main thread
* To run a server without blocking the main thread
* To wait for an anonymous function to complete or give an error
* To set a timeout for a function
* To allow an object's run function to run concurrently with the main thread

### Micro

A text editor.

Uses of concurrency in Micro:

* To run a command without blocking the main thread
* To separate tasks through use of channels
* To do IO without blocking the main thread
* To do multiple queries concurrently
* To set a timeout for a function