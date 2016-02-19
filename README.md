community
===
An approach to developing simplified API(s) as a boilerplate in Golang leveraging MongoDB and JSON-API.

#### why?
I have spent the last sixteen months struggling to develop services using Golang in a way that seems successful.  Having some measureable amount of success or being successful includes having repeatable steps and a foundation for progress.  I have developed, refactored, deleted, and started over more times than I can count and now I am committing myself to going in one direction, forward.

#### approach
Each major commit or branch of this code will represent some lesson or lessons I am trying to learn.  They will likely be formalized into more of an article at some point for me to declare some small amount of victory with this approach, but we shall see.

#### inspiration
I continually monitor the progression of great code coming out of my friends and colleagues and I envy their ability to write stuff that doesn't suck.  I am absorbing their brans to help my code not suck nearly as bad.

THanks to all of the various [make a restful api](https://thenewstack.io/make-a-restful-json-api-go/) in golang [articles](https://github.com/ant0ine/go-json-rest) that are [all over the interwebs](https://www.nicolasmerouze.com/how-to-render-json-api-golang-mongodb/) for giving me plenty of information to read.

#### useful links

* [json-api](http://jsonapi.org/) - is where I am trying to get the infos for making this not suck entirely.

## Changelog

#### 'divide' => [0.0.2] - 2016-02-18
Merging the branch 'divide' into master to become version 0.0.2 as unofficially official as I can.  Basically, I decided to spend some time splitting up the single file api into a few more manageable packages.  Not entirely sure they are grouped logically as I still have some Host details in community-api/main.go and such.
- Added 'corsHandler'
- Modify Host data
 
