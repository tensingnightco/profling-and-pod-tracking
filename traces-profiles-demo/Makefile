.PHONY: build push deploy test cleanup

# Variables
IMAGE_NAME = trace-profile-demo
IMAGE_TAG = latest
REGISTRY = docker.io  # Change to your registry

build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

tag:
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

push:
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

deploy:
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml

load-test:
	kubectl apply -f k8s/load-test.yaml

test:
	@echo "Testing connectivity..."
	kubectl run curl-test --image=curlimages/curl --rm -it --restart=Never -- \
		curl -v http://trace-profile-demo:8080/hello

logs:
	kubectl logs -f deployment/trace-profile-demo

status:
	kubectl get pods -l app=trace-profile-demo
	kubectl get svc trace-profile-demo

port-forward:
	kubectl port-forward svc/trace-profile-demo 8080:8080

cleanup:
	kubectl delete -f k8s/load-test.yaml --ignore-not-found
	kubectl delete -f k8s/deployment.yaml --ignore-not-found
	kubectl delete -f k8s/service.yaml --ignore-not-found
	kubectl delete -f k8s/configmap.yaml --ignore-not-found
	kubectl delete pod trace-profile-demo-load-generator --ignore-not-found

full-deploy: build push deploy load-test
	@echo "✅ Full deployment complete!"
	@echo "📊 Check traces in Grafana: http://localhost:3000/explore"
	@echo "🔥 Check profiles in Pyroscope: http://localhost:4040"