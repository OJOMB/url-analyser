![# URL Analyser](public/url-analyser.png)

A simple web app for running some basic analysis on a given URL

Built with React and Golang!

I've prebundled the frontend code, so the program should run on:

```
go build
./url-analyser
```

and be accessible on localhost:8080/

# Please Read

I'm very sorry but I ran out of time with this one. 
I had a very hectic week because of the deteriorating covid situation in the Czech Republic where I currently live. I had to leave the country in a bit of a panic before travel restrictions came into place. Unfortunately as a result I didn't manage to get this project to quite the level I wanted to.

## Issues
* The program works but it has some fairly serious performance issues that stem from the implementation of the method for checking accessibilty of links.
* I didnt manage to get to testing but I would just like to say that I fully realise the importance of testing. I've uploaded a side project of mine which demonstates some Golang testing code I've written recently, the repo can be found [here](https://github.com/OJOMB/trigram-markov-model).
* I hope the frontend code being written in React will not count against me. I realise that it may not have been the most lightweight, efficient choice for this particular use case but I hoped to demonstrate some versatility and show extra skills I might bring to the table.
* Many corner cases are currently unhandled and I didn't get to handling non-200 responses on the frontend

I'm very keen on this position so I hope the shortcomings of this submission will be understood in the context of the chaos going on at the minute

