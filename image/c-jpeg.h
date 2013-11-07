#ifndef _SIMP_C_JPEG_H
#define _SIMP_C_JPEG_H


#include <stdlib.h>
#include <stdbool.h>
#include <stdio.h>
#include <setjmp.h>
#include <jpeglib.h>


typedef struct jpeg_attr * jpeg_attr_ptr;

struct jpeg_attr {
	UINT16 width;
	UINT16 height;
	UINT8  quality;
};

typedef struct jpeg_option  * jpeg_option_ptr;

struct jpeg_option {
	UINT8   quality;
	boolean strip_all;
};

typedef struct my_error_mgr * my_error_ptr;

struct my_error_mgr {
	struct jpeg_error_mgr     pub;
	jmp_buf                   setjmp_buffer;
};

typedef struct _Simp_Image    Simp_Image;

struct _Simp_Image {
	struct my_error_mgr             jerr;
	struct {
		struct jpeg_decompress_struct  ji;
		int                            w, h, q;
		FILE                          *f;
	} in;
	struct {
		struct jpeg_compress_struct    ji;
		FILE                          *f;
		struct {
			unsigned char           **data;
			int                      *size;
		} mem;
	} out;
	struct jpeg_option              wopt;
	JSAMPARRAY 					 	buf; // ptr for unsigned char **lines
};

Simp_Image  *simp_open_stdio 	(FILE *infile);
Simp_Image  *simp_open_mem	 	(unsigned char *data, unsigned int size);
void         simp_close      	(Simp_Image *im);
bool         simp_output_file   (Simp_Image *im, FILE *outfile);
bool         simp_output_mem    (Simp_Image *im, unsigned char **data, unsigned long *size);
int 		 estimate_jpeg_quality(j_decompress_ptr jpeg_info);
int          read_jpeg_file (FILE * infile, jpeg_attr_ptr ia_ptr);
int          write_jpeg_file (FILE * infile, FILE * outfile, jpeg_option_ptr option);
int 		 simp_get_width     (Simp_Image *im);
int 		 simp_get_height    (Simp_Image *im);
int 		 simp_get_quality   (Simp_Image *im);
void 		 simp_set_quality   (Simp_Image *im, int quality);

#endif
