goose-up:
	cd sql/schema && goose postgres postgres://postgres:postgres@localhost:5432/chirpy up

goose-down:
	cd sql/schema && goose postgres postgres://postgres:postgres@localhost:5432/chirpy down

goose-re: goose-down goose-up
