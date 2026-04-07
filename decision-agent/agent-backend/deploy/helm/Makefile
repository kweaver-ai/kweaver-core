test:
	helm template agent-factory agent-backend>./helm-template-gen.yaml

test-business-domain:
	./test-business-domain-toggle.sh

test-split:
	helm template agent-factory agent-backend>./helm-template-gen.yaml && \
	mkdir -p ./helm-template-gen-split && \
	cd ./helm-template-gen-split && \
	rm -f *.yaml && \
	awk '/^---/ {if (NR>1) close(outfile); outfile="split-"++i".yaml"} {print > outfile}' ../helm-template-gen.yaml && \
	for f in split-*.yaml; do \
		grep -E "^# Source: " "$$f" | head -1 | sed 's/# Source: agent-backend\/templates\///' | sed 's/.yaml$$//' | xargs -I {} mv "$$f" "{}.yaml"; \
	done

lint:
	cd agent-backend && helm lint .
