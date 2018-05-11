all: dev demo prod
	
dev:
	apex -r ap-southeast-1 --env dev deploy
	apex -r ap-southeast-1 --env dev invoke simple < event.json

demo:
	apex -r ap-southeast-1 --env demo deploy
	apex -r ap-southeast-1 --env demo invoke simple < event.json

prod:
	apex -r ap-southeast-1 --env prod deploy
	apex -r ap-southeast-1 --env prod invoke simple < event.json


.PHONY: dev demo prod
