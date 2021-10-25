release:
	goreleaser --rm-dist

pprof-profile:
	ssh sso-kw1-amzn2.nikkansports.com curl -s localhost:8181/debug/pprof/profile <<<"" | gzip -c > pprof.zunproxy.sso-kw1-amzn2.nikkansports.com."$(shell date +%Y%m%d-%H%M%S)".pb.gz