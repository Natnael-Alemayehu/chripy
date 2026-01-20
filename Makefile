goose-up:
	cd sql/schema && goose postgres postgres://postgres:postgres@localhost:5432/chirpy up

goose-down:
	cd sql/schema && goose postgres postgres://postgres:postgres@localhost:5432/chirpy down

goose-re: goose_up goose_down
