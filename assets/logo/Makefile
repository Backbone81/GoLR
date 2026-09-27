# Produces the logo in the formats and sizes the IDE plugins need from the originals logo_light.svg and logo_dark.svg.
# Requires sed, rsvg-convert and exiftool (Ubuntu: apt install librsvg2-bin libimage-exiftool-perl).

# Removes a PNG whose recipe failed, so a PNG which was rendered but not stripped never counts as up to date.
.DELETE_ON_ERROR:

VARIANTS := light dark
SVG_SIZES := 16 40 80
PNG_SIZES := 32 256

SVGS := $(foreach v,$(VARIANTS),$(foreach s,$(SVG_SIZES),logo_$(v)_$(s).svg))
PNGS := $(foreach v,$(VARIANTS),$(foreach s,$(PNG_SIZES),logo_$(v)_$(s).png))

INTELLIJ_RESOURCES := ../../ide/intellij/src/main/resources
VSCODE_IMAGES := ../../ide/vscode/images

.PHONY: all
all: $(SVGS) $(PNGS)

# The size only changes the width and height of the root svg element; the viewBox keeps the drawing scaled to fit.
logo_light_%.svg: logo_light.svg
	sed -E '/<svg /s/ width="[^"]*" height="[^"]*"/ width="$*" height="$*"/' $< > $@

logo_dark_%.svg: logo_dark.svg
	sed -E '/<svg /s/ width="[^"]*" height="[^"]*"/ width="$*" height="$*"/' $< > $@

# exiftool strips all metadata, so nothing about the machine or the time of the build ends up in the PNGs, whatever
# rsvg-convert writes by default. PNGs are transparent unless BACKGROUND is set for them.
logo_light_%.png: logo_light.svg
	rsvg-convert --width $* --height $* $(if $(BACKGROUND),--background-color $(BACKGROUND)) --output $@ $<
	exiftool -quiet -all= -overwrite_original $@

logo_dark_%.png: logo_dark.svg
	rsvg-convert --width $* --height $* $(if $(BACKGROUND),--background-color $(BACKGROUND)) --output $@ $<
	exiftool -quiet -all= -overwrite_original $@

# The VS Code extension icon has no dark variant, so it gets a white background to keep its contrast on dark themes.
logo_light_256.png: BACKGROUND := white

# Copies the logo into the IDE plugins under the file names each IDE expects.
.PHONY: ide
ide: all
	cp logo_light_40.svg $(INTELLIJ_RESOURCES)/META-INF/pluginIcon.svg
	cp logo_dark_40.svg $(INTELLIJ_RESOURCES)/META-INF/pluginIcon_dark.svg
	cp logo_light_16.svg $(INTELLIJ_RESOURCES)/icons/golr.svg
	cp logo_dark_16.svg $(INTELLIJ_RESOURCES)/icons/golr_dark.svg
	cp logo_light_32.png $(VSCODE_IMAGES)/golr.png
	cp logo_dark_32.png $(VSCODE_IMAGES)/golr_dark.png
	cp logo_light_256.png $(VSCODE_IMAGES)/icon.png
