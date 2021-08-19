build:
	docker rmi webhooktest:0.1.0 || true
	docker build -t webhooktest:0.1.0 .

pushimage:
	kind load docker-image "webhooktest:0.1.0"

deploy:
	kubectl delete deploy -l app=inovex-webhook
	kubectl apply -f deployment.yml

clean:
	kubectl delete -f deployment.yml
	kubectl delete -f test_deployment.yml

test:
	kubectl delete -f test_deployment.yml || true
	kubectl apply -f test_deployment.yml
