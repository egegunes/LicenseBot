build:
	docker build -t eu.gcr.io/kubernetes-222419/licensebot .

push:
	docker push eu.gcr.io/kubernetes-222419/licensebot
