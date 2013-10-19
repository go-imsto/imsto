#include <stdlib.h>
#include <stdio.h>
#include <jpeglib.h>
#include <setjmp.h>
#include "c-jpeg.h"
#include "_cgo_export.h"

struct my_error_mgr {
  struct jpeg_error_mgr pub;

  jmp_buf setjmp_buffer;
};

typedef struct my_error_mgr * my_error_ptr;

void my_error_exit (j_common_ptr cinfo)
{
  my_error_ptr myerr = (my_error_ptr) cinfo->err;
  (*cinfo->err->output_message) (cinfo);
  longjmp(myerr->setjmp_buffer, 1);
}

int ReadJPEGFile (FILE * infile, jpeg_attr_ptr ia_ptr)
{
  // printf("ReadJPEGFile %s\n", filename);
  struct jpeg_decompress_struct cinfo;

  struct my_error_mgr jerr;
  // FILE * infile;
  JSAMPARRAY buffer;
  int row_stride;
  //jpeg_attr attr;
  UINT8 quality;

  // if ((infile = fopen(filename, "rb")) == NULL) {
  //   fprintf(stderr, "can't open %s\n", filename);
  //   return 0;
  // }

  cinfo.err = jpeg_std_error(&jerr.pub);
  jerr.pub.error_exit = my_error_exit;

  if (setjmp(jerr.setjmp_buffer)) {

    jpeg_destroy_decompress(&cinfo);
    // fclose(infile);
    return 0;
  }

  jpeg_create_decompress(&cinfo);

  jpeg_stdio_src(&cinfo, infile);
  jpeg_read_header(&cinfo, TRUE);

#ifdef IM_DEBUG
  printf("jpeg header: %dx%d %dbit %c\n", cinfo.image_width, cinfo.image_height,
    cinfo.num_components*8, (cinfo.progressive_mode?'P':'N'));
#endif

  quality = EstimateJPEGQuality(&cinfo);
#ifdef IM_DEBUG
  printf("quality: %d\n", quality);
#endif

  ia_ptr->width = (UINT16)cinfo.image_width;
  ia_ptr->height = (UINT16)cinfo.image_height;
  ia_ptr->quality = (UINT8)quality;
  // ia_ptr = jpeg_attr{(UINT16)cinfo.image_width, (UINT16)cinfo.image_height, (UINT8)quality}

  (void) jpeg_start_decompress(&cinfo);
/*
  row_stride = cinfo.output_width * cinfo.output_components;

  buffer = (*cinfo.mem->alloc_sarray)
		((j_common_ptr) &cinfo, JPOOL_IMAGE, row_stride, 1);

  while (cinfo.output_scanline < cinfo.output_height) {

    (void) jpeg_read_scanlines(&cinfo, buffer, 1);

    //put_scanline_someplace(buffer[0], row_stride);
  }
*/
  (void) jpeg_finish_decompress(&cinfo);

  jpeg_destroy_decompress(&cinfo);

  // fclose(infile);

  return 1;
}

int WriteJPEGFile (FILE * infile, FILE * outfile, jpeg_option_ptr option)
{
	int j;
  struct jpeg_decompress_struct dinfo;
  struct jpeg_compress_struct cinfo;
  struct my_error_mgr jcerr,jderr;

  /* initialize decompression object */
  dinfo.err = jpeg_std_error(&jderr.pub);
  jpeg_create_decompress(&dinfo);
  jderr.pub.error_exit=my_error_exit;
  // jderr.pub.output_message=my_output_message;

  /* initialize compression object */
  cinfo.err = jpeg_std_error(&jcerr.pub);
  jpeg_create_compress(&cinfo);
  jcerr.pub.error_exit=my_error_exit;
  // jcerr.pub.output_message=my_output_message;

  JSAMPARRAY buf = NULL;

   /* setup error handling for decompress */
   if (setjmp(jderr.setjmp_buffer)) {
      jpeg_abort_decompress(&dinfo);
      // fclose(infile);
      if (buf) {
	for (j=0;j<dinfo.output_height;j++) free(buf[j]);
	free(buf); buf=NULL;
      }
      // if (!quiet_mode) printf(" [ERROR]\n");
      return 0;
   }

   jpeg_stdio_src(&dinfo, infile);
   jpeg_read_header(&dinfo, TRUE); 

     jpeg_start_decompress(&dinfo);

     buf = malloc(sizeof(JSAMPROW)*dinfo.output_height);
     // if (!buf) fatal("not enough memory");
     for (j=0;j<dinfo.output_height;j++) {
       buf[j]=malloc(sizeof(JSAMPLE)*dinfo.output_width*
		     dinfo.out_color_components);
       // if (!buf[j]) fatal("not enough memory");
     }

     while (dinfo.output_scanline < dinfo.output_height) {
       jpeg_read_scanlines(&dinfo,&buf[dinfo.output_scanline],
			   dinfo.output_height-dinfo.output_scanline);
     }

   if (setjmp(jcerr.setjmp_buffer)) {
      jpeg_abort_compress(&cinfo);
      jpeg_abort_decompress(&dinfo);
      // fclose(outfile);
      // outfile=NULL;
      // if (infile) fclose(infile);
      // if (!quiet_mode) printf(" [Compress ERROR]\n");
      if (buf) {
	for (j=0;j<dinfo.output_height;j++) free(buf[j]);
	free(buf); buf=NULL;
      }
      return 0;
   }

   jpeg_stdio_dest(&cinfo, outfile);


     cinfo.in_color_space=dinfo.out_color_space;
     cinfo.input_components=dinfo.output_components;
     cinfo.image_width=dinfo.image_width;
     cinfo.image_height=dinfo.image_height;
     jpeg_set_defaults(&cinfo); 
     jpeg_set_quality(&cinfo,(int)option->quality,TRUE);
     if ( /*(*/dinfo.progressive_mode /*|| all_progressive) && !all_normal*/ )
       jpeg_simple_progression(&cinfo);
     cinfo.optimize_coding = TRUE;

     j=0;
     jpeg_start_compress(&cinfo,TRUE);
     
     /* write image */
     while (cinfo.next_scanline < cinfo.image_height) {
       jpeg_write_scanlines(&cinfo,&buf[cinfo.next_scanline],
			    dinfo.output_height);
     }

   jpeg_finish_compress(&cinfo);
   fflush(outfile);

   if (buf) {
     for (j=0;j<dinfo.output_height;j++) free(buf[j]);
     free(buf); buf=NULL;
   }
   jpeg_finish_decompress(&dinfo);
   // fclose(infile);
   // fclose(outfile);
   outfile=NULL;

  jpeg_destroy_decompress(&dinfo);
  jpeg_destroy_compress(&cinfo);

	return 1;
}

int EstimateJPEGQuality(j_decompress_ptr jpeg_info)
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
		    printf("Quality: %d (approximate)\n", save_quality);
		  else
		    printf("Quality: %d\n", save_quality);
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
