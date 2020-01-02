DOT_FILES=$(wildcard *.dot)
DOTPNG=$(DOT_FILES:.dot=.svg)

.PHONY: all
all: $(DOTPNG)

%.svg: %.dot
	dot -Tsvg $< > $@
