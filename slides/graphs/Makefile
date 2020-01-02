DOT_FILES=$(wildcard *.dot)
DOTSVG=$(DOT_FILES:.dot=.svg)
DOTPNG=$(DOT_FILES:.dot=.png)

.PHONY: all
all: $(DOTSVG)

%.svg: %.dot
	dot -Tsvg $< > $@
	./base64ify $@
	mv $@ ../

%.png: %.dot
	dot -Tpng $< > $@
