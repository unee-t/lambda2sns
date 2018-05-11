all: dev demo prod

dev:
	apex -r ap-southeast-1 --env dev deploy

demo:
	apex -r ap-southeast-1 --env demo deploy

prod:
	apex -r ap-southeast-1 --env prod deploy

testdev:
	apex -r ap-southeast-1 --env dev invoke simple < event.json

testdemo:
	apex -r ap-southeast-1 --env demo invoke simple < event.json

testprod:
	apex -r ap-southeast-1 --env prod invoke simple < event.json


.PHONY: dev demo prod testdev testdemo testprod
