
#include <stdlib.h>
#include <stdio.h>
#include <jpeglib.h>
#include <setjmp.h>
#include "c-jpeg.h"
// #include "_cgo_export.h"

void 		_simp_init      (Simp_Image *im);
Simp_Image *_simp_read_head (Simp_Image *im);

#if defined IM_DEBUG && defined JPEG_DEBUG
static void 
my_message_handle(j_common_ptr cinfo, int msg_level)
{
	char message[JMSG_LENGTH_MAX];
	struct jpeg_error_mgr *err;

	message[0]='\0';
	err=cinfo->err;
	(err->format_message)(cinfo, message);

	printf(">\t%s\n", message);
}
#endif

static void 
my_error_exit (j_common_ptr cinfo)
{
	my_error_ptr myerr = (my_error_ptr) cinfo->err;
	(*cinfo->err->output_message) (cinfo);
	longjmp(myerr->setjmp_buffer, 1);
	return;
}


Simp_Image *
simp_open_stdio(FILE *infile)
{
#ifdef IM_DEBUG
	printf("[start simp_open_stdio]\n");
#endif

	Simp_Image *im;
	im = calloc(1, sizeof(Simp_Image));
	
	if (infile == NULL) {
		fprintf(stderr, "infile is null\n");
		return NULL;
	}

#ifdef IM_DEBUG
	printf("start set in.f\n");
#endif
	im->in.f = infile;
	if (!im->in.f) {
		simp_close(im);
		return NULL;
	}

	_simp_init(im);

	if (im->in.f) jpeg_stdio_src(&(im->in.ji), infile);

	return _simp_read_head(im);
}

Simp_Image *
simp_open_mem(unsigned char *data, unsigned int size)
{
#ifdef IM_DEBUG
	printf("[start simp_open_mem: %d]\n", size);
#endif

	Simp_Image *im;
	im = calloc(1, sizeof(Simp_Image));
	_simp_init(im);
	jpeg_mem_src(&(im->in.ji), data, (unsigned long)size);
	return _simp_read_head(im);
}

void
_simp_init(Simp_Image *im)
{
#ifdef IM_DEBUG
	printf("[start simp_init]\n");
#endif

	im->in.ji.err = jpeg_std_error(&(im->jerr.pub));
#if defined IM_DEBUG && defined JPEG_DEBUG
	im->jerr.pub.emit_message = my_message_handle;
#endif
	im->jerr.pub.error_exit = my_error_exit;
	
	if (setjmp(im->jerr.setjmp_buffer)) {
		// error:
		simp_close(im);
		im = NULL;
		fprintf(stderr, "error\n");
	}

	jpeg_create_decompress(&(im->in.ji));

}


Simp_Image *
_simp_read_head(Simp_Image *im)
{
#ifdef IM_DEBUG
	printf("[start simp_read_head]\n");
#endif

	int rc;
	rc = jpeg_read_header(&(im->in.ji), TRUE);

	if (rc != 1) {
		fprintf(stderr, "File does not seem to be a normal JPEG");
		simp_close(im);
		im = NULL;
		return NULL;
	}

	im->in.w = im->in.ji.image_width;
	im->in.h = im->in.ji.image_height;
	if (im->in.w <= 1 || im->in.h <= 1) {
		simp_close(im);
		im = NULL;
		return NULL;
	}

#ifdef IM_DEBUG
	printf("[jpeg header: %dx%d %dbit %c]\n", im->in.w, im->in.h,
		im->in.ji.num_components*8, (im->in.ji.progressive_mode?'P':'N'));
#endif

	im->in.q = estimate_jpeg_quality(&(im->in.ji));
	im->wopt.quality = (UINT8)im->in.q;
#ifdef IM_DEBUG
	printf("[quality: %d]\n", im->in.q);
#endif

	return im;
}

int
simp_get_width(Simp_Image *im)
{
	return im->in.w;
}

int
simp_get_height(Simp_Image *im)
{
	return im->in.h;
}

int
simp_get_quality(Simp_Image *im)
{
	return im->in.q;
}

void
simp_set_quality(Simp_Image *im, int quality)
{
	if (quality > 0 && quality <= 100) {
		im->wopt.quality = quality;
	}
#ifdef IM_DEBUG
	else
		printf("[invalid quality: %d]\n", quality);
#endif

}

void
simp_close(Simp_Image *im)
{
	if (im->in.f)      jpeg_destroy_decompress(&(im->in.ji));
	if (im->in.f)      fclose(im->in.f);
	if (im->buf)       free(im->buf);
	if (im->out.f)     jpeg_destroy_compress(&(im->out.ji));
	if (im->out.f)     fclose(im->out.f);
	free(im);
}

static bool
_simp_decode(Simp_Image *im)
{
#ifdef IM_DEBUG
	printf("[start simp_decode]\n");
#endif
	// JSAMPARRAY buf = NULL;
	int j;

	/* setup error handling for decompress */
	if (setjmp(im->jerr.setjmp_buffer)) {
		jpeg_abort_decompress(&(im->in.ji));
		// fclose(infile);
		if (im->buf) {
			for (j=0;j<im->in.ji.output_height;j++) free(im->buf[j]);
			free(im->buf); im->buf=NULL;
		}
		simp_close(im);
#ifdef IM_DEBUG
	printf("[simp decode ERROR]\n");
#endif

		return FALSE;
	}

	jpeg_start_decompress(&(im->in.ji));

	im->buf = malloc(sizeof(JSAMPROW)*im->in.ji.output_height);
	// if (!im->buf) fatal("not enough memory");
	for (j=0;j<im->in.ji.output_height;j++) {
		im->buf[j]=malloc(sizeof(JSAMPLE)*im->in.ji.output_width*
			im->in.ji.out_color_components);
		// if (!im->buf[j]) fatal("not enough memory");
	}

	while (im->in.ji.output_scanline < im->in.ji.output_height) {
		jpeg_read_scanlines(&(im->in.ji),&im->buf[im->in.ji.output_scanline],
			im->in.ji.output_height-im->in.ji.output_scanline);
	}

	if (setjmp(im->jerr.setjmp_buffer)) {
		jpeg_abort_decompress(&(im->in.ji));
#ifdef IM_DEBUG
	printf("[Decompress ERROR]\n");
#endif

		if (im->buf) {
			for (j=0;j<im->in.ji.output_height;j++) free(im->buf[j]);
			free(im->buf); im->buf=NULL;
		}
		simp_close(im);
		return 0;
	}
	jpeg_finish_decompress(&(im->in.ji));
#ifdef IM_DEBUG
	printf("[decode OK %d lines]\n", j);
#endif
	return true;
}

static bool
_simp_encode(Simp_Image *im)
{
#ifdef IM_DEBUG
	printf("[start simp_encode]\n");
#endif

	void  *data = NULL;
	size_t size = 0;

	im->out.ji.err = jpeg_std_error(&(im->jerr.pub));
	im->jerr.pub.error_exit = my_error_exit;
	if (setjmp(im->jerr.setjmp_buffer)) return FALSE;
	jpeg_create_compress(&(im->out.ji));

	return TRUE;
}

static bool
_simp_write(Simp_Image *im)
{
#ifdef IM_DEBUG
	printf("[start simp_write]\n");
#endif

	int j;
	im->out.ji.in_color_space=im->in.ji.out_color_space;
	im->out.ji.input_components=im->in.ji.output_components;
	im->out.ji.image_width=im->in.ji.image_width;
	im->out.ji.image_height=im->in.ji.image_height;
	jpeg_set_defaults(&(im->out.ji)); 
	if (im->wopt.quality > 0 && im->wopt.quality < 100) {
#ifdef IM_DEBUG
	printf("[set wopt quality %d]\n", im->wopt.quality);
#endif
		jpeg_set_quality(&(im->out.ji),(int)im->wopt.quality,TRUE);
	}
	else
		jpeg_set_quality(&(im->out.ji),(int)im->in.q,TRUE);

	if ( /*(*/im->in.ji.progressive_mode /*|| all_progressive) && !all_normal*/ )
		jpeg_simple_progression(&(im->out.ji));
	im->out.ji.optimize_coding = TRUE;

	j=0;
	jpeg_start_compress(&(im->out.ji),TRUE);
	
	/* write image */
	while (im->out.ji.next_scanline < im->out.ji.image_height) {
		jpeg_write_scanlines(&(im->out.ji),&im->buf[im->out.ji.next_scanline],
				im->in.ji.output_height);
	}

	jpeg_finish_compress(&(im->out.ji));

	if (im->buf) {
		for (j=0;j<im->in.ji.output_height;j++) free(im->buf[j]);
		free(im->buf); im->buf=NULL;
	}

	jpeg_destroy_decompress(&im->in.ji);
	jpeg_destroy_compress(&(im->out.ji));
#ifdef IM_DEBUG
	printf("[write OK %d lines]\n", j);
#endif

	return TRUE;
}

bool
simp_output_file (Simp_Image *im, FILE *outfile)
{

	if (outfile == NULL) {
		fprintf(stderr, "infile is null\n");
		return FALSE;
	}

	if (!_simp_decode(im))
		return FALSE;
	if (!_simp_encode(im))
		return FALSE;
	jpeg_stdio_dest(&(im->out.ji), outfile);

	bool r = _simp_write(im);
	fflush(outfile);
	return r;
}


bool
simp_output_mem  (Simp_Image *im, unsigned char **data, unsigned long *size)
{

	if (!_simp_decode(im))
		return FALSE;
	if (!_simp_encode(im))
		return FALSE;
	jpeg_mem_dest(&(im->out.ji), data, size);

	bool r = _simp_write(im);
	return r;
}


int read_jpeg_file (FILE * infile, jpeg_attr_ptr ia_ptr)
{
	Simp_Image *im;
	// UINT8 quality;
	im = simp_open_stdio(infile);

	if (im == NULL) {
		fprintf(stderr, "simp_open_stdio failed\n");
		return 0;
	}

	ia_ptr->width = (UINT16)im->in.w;
	ia_ptr->height = (UINT16)im->in.h;
	ia_ptr->quality = (UINT8)im->in.q;
	simp_close(im);

	// fclose(infile);

	return 1;
}

int write_jpeg_file (FILE * infile, FILE * outfile, jpeg_option_ptr option)
{
	if (infile == NULL)
	{
		fprintf(stderr, "input file error\n");
		return 0;
	}
	if (outfile == NULL)
	{
		fprintf(stderr, "output file error\n");
		return 0;
	}
	Simp_Image *im;
	im = simp_open_stdio(infile);
	// im->wopt.strip_all = option.strip_all;
	im->wopt.quality = option->quality;
	simp_output_file(im, outfile);
	simp_close(im);

	return 1;
}

int estimate_jpeg_quality(j_decompress_ptr jpeg_info)
{
	int save_quality;
	register long i;
	save_quality=0;
	int hashval, sum;

	sum=0;
	for (i=0; i < NUM_QUANT_TBLS; i++)
	{
		int
			j;

		if (jpeg_info->quant_tbl_ptrs[i] != NULL)
			for (j=0; j < DCTSIZE2; j++)
				{
		UINT16 *c;
		c=jpeg_info->quant_tbl_ptrs[i]->quantval;
		sum+=c[j];
				}
	}
			if ((jpeg_info->quant_tbl_ptrs[0] != NULL) &&
		(jpeg_info->quant_tbl_ptrs[1] != NULL))
	{
		int
			hash[] =
			{
				1020, 1015,  932,  848,  780,  735,  702,  679,  660,  645,
				632,  623,  613,  607,  600,  594,  589,  585,  581,  571,
				555,  542,  529,  514,  494,  474,  457,  439,  424,  410,
				397,  386,  373,  364,  351,  341,  334,  324,  317,  309,
				299,  294,  287,  279,  274,  267,  262,  257,  251,  247,
				243,  237,  232,  227,  222,  217,  213,  207,  202,  198,
				192,  188,  183,  177,  173,  168,  163,  157,  153,  148,
				143,  139,  132,  128,  125,  119,  115,  108,  104,   99,
				94,   90,   84,   79,   74,   70,   64,   59,   55,   49,
				45,   40,   34,   30,   25,   20,   15,   11,    6,    4,
				0
			};

		int
			sums[] =
			{
				32640,32635,32266,31495,30665,29804,29146,28599,28104,27670,
				27225,26725,26210,25716,25240,24789,24373,23946,23572,22846,
				21801,20842,19949,19121,18386,17651,16998,16349,15800,15247,
				14783,14321,13859,13535,13081,12702,12423,12056,11779,11513,
				11135,10955,10676,10392,10208, 9928, 9747, 9564, 9369, 9193,
				9017, 8822, 8639, 8458, 8270, 8084, 7896, 7710, 7527, 7347,
				7156, 6977, 6788, 6607, 6422, 6236, 6054, 5867, 5684, 5495,
				5305, 5128, 4945, 4751, 4638, 4442, 4248, 4065, 3888, 3698,
				3509, 3326, 3139, 2957, 2775, 2586, 2405, 2216, 2037, 1846,
				1666, 1483, 1297, 1109,  927,  735,  554,  375,  201,  128,
				0
			};

		hashval=(jpeg_info->quant_tbl_ptrs[0]->quantval[2]+
			jpeg_info->quant_tbl_ptrs[0]->quantval[53]+
			jpeg_info->quant_tbl_ptrs[1]->quantval[0]+
			jpeg_info->quant_tbl_ptrs[1]->quantval[DCTSIZE2-1]);
		for (i=0; i < 100; i++)
			{
				if ((hashval >= hash[i]) || (sum >= sums[i]))
		{
			save_quality=i+1;
			#ifdef IM_DEBUG
			if ((hashval > hash[i]) || (sum > sums[i]))
				printf("[Quality: %d (approximate)]\n", save_quality);
			else
				printf("[Quality: %d]\n", save_quality);
			#endif
			break;
		}
			}
	}
			else
	if (jpeg_info->quant_tbl_ptrs[0] != NULL)
		{
			int
				bwhash[] =
				{
		510,  505,  422,  380,  355,  338,  326,  318,  311,  305,
		300,  297,  293,  291,  288,  286,  284,  283,  281,  280,
		279,  278,  277,  273,  262,  251,  243,  233,  225,  218,
		211,  205,  198,  193,  186,  181,  177,  172,  168,  164,
		158,  156,  152,  148,  145,  142,  139,  136,  133,  131,
		129,  126,  123,  120,  118,  115,  113,  110,  107,  105,
		102,  100,   97,   94,   92,   89,   87,   83,   81,   79,
		76,   74,   70,   68,   66,   63,   61,   57,   55,   52,
		50,   48,   44,   42,   39,   37,   34,   31,   29,   26,
		24,   21,   18,   16,   13,   11,    8,    6,    3,    2,
		0
				};

			int
				bwsum[] =
				{
		16320,16315,15946,15277,14655,14073,13623,13230,12859,12560,
		12240,11861,11456,11081,10714,10360,10027, 9679, 9368, 9056,
		8680, 8331, 7995, 7668, 7376, 7084, 6823, 6562, 6345, 6125,
		5939, 5756, 5571, 5421, 5240, 5086, 4976, 4829, 4719, 4616,
		4463, 4393, 4280, 4166, 4092, 3980, 3909, 3835, 3755, 3688,
		3621, 3541, 3467, 3396, 3323, 3247, 3170, 3096, 3021, 2952,
		2874, 2804, 2727, 2657, 2583, 2509, 2437, 2362, 2290, 2211,
		2136, 2068, 1996, 1915, 1858, 1773, 1692, 1620, 1552, 1477,
		1398, 1326, 1251, 1179, 1109, 1031,  961,  884,  814,  736,
		667,  592,  518,  441,  369,  292,  221,  151,   86,   64,
		0
				};

			hashval=(jpeg_info->quant_tbl_ptrs[0]->quantval[2]+
				jpeg_info->quant_tbl_ptrs[0]->quantval[53]);
			for (i=0; i < 100; i++)
				{
		if ((hashval >= bwhash[i]) || (sum >= bwsum[i]))
			{
				save_quality=i+1;
				#ifdef IM_DEBUG
				if ((hashval > bwhash[i]) || (sum > bwsum[i]))
					printf("Quality: %ld (approximate)\n", i+1);
				else
					printf("Quality: %ld\n", i+1);
				#endif
				break;
			}
				}
		}

	return save_quality;
}
