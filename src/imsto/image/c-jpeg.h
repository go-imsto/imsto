
#include <stdlib.h>
#include <stdio.h>
#include <jpeglib.h>


typedef struct jpeg_attr * jpeg_attr_ptr;

struct jpeg_attr {
	UINT16 width;
	UINT16 height;
	UINT8  quality;
};

typedef struct jpeg_option * jpeg_option_ptr;

struct jpeg_option {
	UINT8   quality;
	boolean strip_all;
};

extern int EstimateJPEGQuality(j_decompress_ptr jpeg_info);

extern int ReadJPEGFile (FILE * infile, jpeg_attr_ptr ia_ptr);

extern int WriteJPEGFile (FILE * infile, FILE * outfile, jpeg_option_ptr option);
