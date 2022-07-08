# Time Capsule
## A project I built while learning the basics of Blockchain technology.
### This is basically a time capsule running on a locally hosted Blockchain

- Basic web time capsule
- Uses AFMA protocol to reach consensus
- ✨ Blockchain Magic ✨ uses blockchain tech to built chain of text messages.

Thanks for checking it out. 


## To Run --- Mac Only
Run the server.go and the only parameter to pass is n (total number of processes which needs to match the n when running peer.go)
```
go run server.go -n="numOfProcesses"
```

once the server is up you can run the apple script to boot up the peers they will decide on a first block and then you can send text over from the server the front end will always be one block behind do to me not rerendering, if you want to see the block that got picked from the message that you typed either reload the page or just type another message in to the field and it will auto refresh when you click enter(the button or the key)

for some reason sometimes my server doesn't grab the value from the post and I couldn't figure out why it works until it doesn't ususally at the fifth block.
