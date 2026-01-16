SERVICES := orders payments

.PHONY: db-create migrate-up migrate-down

db-create:
	@for svc in $(SERVICES); do \
		$(MAKE) -C $$svc db-create; \
	done

migrate-up:
	@for svc in $(SERVICES); do \
		$(MAKE) -C $$svc migrate-up; \
	done

migrate-down:
	@for svc in $(SERVICES); do \
		$(MAKE) -C $$svc migrate-down; \
	done
