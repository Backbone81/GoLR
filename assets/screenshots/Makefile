# Strips all metadata in place from the hand-taken PNG screenshots in this folder, like the capture time and the
# software which took them.
# Requires exiftool (Ubuntu: apt install libimage-exiftool-perl).

# exiftool walks the folder itself, as screenshot file names often contain spaces.
.PHONY: all
all:
	exiftool -quiet -recurse -ext png -all= -overwrite_original .
