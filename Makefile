
.PHONY: run
run:
	operator-sdk up local --namespace="" --operator-flags="--zap-encoder=console"
