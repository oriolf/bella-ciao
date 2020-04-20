Bella Ciao
==========

Bella Ciao is a software for managing primaries voting, aiming for simplicity and ease of use. It is not well suited for long 
term decision making of large organizations, instead it is better suited for one shot primaries of small groups which do not 
have the resources to buy a more complete solution.


Design decisions
----------------

In order to achieve the ease of use, a number of design decisions have been made that simplify the software and makes it easier
to install and use, but also makes it less adequate for larger use cases. Some of these decisions are:

- The software and solutions used are the most lightweight possible: no production scale databases, instead SQLite; no production
  scale web servers, instead the Go language HTTP Server; no email server, but you can specify one if you have it.
- Just one vote per deployment. Even though it sounds extreme, most organizations do their primaries once every few years, then 
  forget about it and have to begin from scratch again the next time. This approach works better, since the effort is way less.
- No different zones for electors, everyone who can vote is considered the same. Zones have to be configured and require much more
  effort although they are rarely used.
