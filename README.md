	NAME
		packchans - pack grayscale images into RGB

	SYNOPSIS
		packchans -o orm.tga occlusion.tga roughness.tga metallic.tga

	DESCRIPTION
		Packchans takes 3 input images and packs their intensities,
		in red-green-blue order, into the channels of an RGB image. The
		output is a 24-bit TARGA file.

	OPTIONS
		-o file
		       output filename (default packed.tga)
