

(function($){

	// image drop box support

	jQuery.event.props.push("dataTransfer");

	var opts = {},
		default_opts = {
			url: '',
			refresh: 1000,
			field_id: '',
			field_name: 'userfile',
			maxfiles: 25,		   // Ignored if queuefiles is set > 0
			maxfilesize: 1,		 // MB file size limit
			queuefiles: 0,		  // Max files before queueing (for large volume uploads)
			queuewait: 200,		 // Queue wait time if full
			data: {},
			headers: {},
			drop: empty,
			dragEnter: empty,
			dragOver: empty,
			dragLeave: empty,
			docEnter: empty,
			docOver: empty,
			docLeave: empty,
			beforeEach: empty,
			afterAll: empty,
			rename: empty,
			error: function(err, file, i) {
				alert(err);
			},
			imageCreated: empty,
			imageLoaded: empty,
			uploadStarted: empty,
			uploadFinished: empty,
			progressUpdated: empty,
			speedUpdated: empty
		},
		errors = ["BrowserNotSupported", "TooManyFiles", "FileTooLarge", "UnselectFile"],
		doc_leave_timer, stop_loop = false,
		images,
		files_count = 0,
		files;

	$.fn.imgdrop = function(options) {
		opts = $.extend( {}, default_opts, options );
		// 经测试发现：Firefox 13 dragLeave 事件不支持了, 貌似从4开始就不支持了

		this.bind('drop', drop).bind('dragenter', dragEnter).bind('dragover', dragOver).bind('dragleave', dragLeave);
		$(document).bind('drop', docDrop).bind('dragenter', docEnter).bind('dragover', docOver).bind('dragleave', docLeave);
		$('#' + opts.field_id).change(function(e) {
			opts.drop(e);
			files = e.target.files;
			files_count = files.length;
			preview(files);
		});
	};

	function drop(e) {//log('drop');
		opts.drop(e);
		files = e.dataTransfer.files;
		files_count = files.length;
		//upload();
		preview(files);
		e.preventDefault();
		return false;
	}

	function preview (files) {
		if (files === null || files === undefined) {
			opts.error(errors[0]);
			return false;
		}

		stop_loop = false;
		if (!files) {
			opts.error(errors[3]);
			return false;
		}

		files_count = files.length;
		for (var i=0; i<files_count; i++) {
			if (stop_loop) return false;
			createImage(files[i]);
		}
	}

	function prettySize(bytes){	// simple function to show a friendly size
		var i = 0;
		while(1023 < bytes){
			bytes /= 1024;
			++i;
		};
		return  i ? bytes.toFixed(2) + ["", " Kb", " Mb", " Gb", " Tb"][i] : bytes + " bytes";
	}

	function createImage (file) {
		var imageType = /image.*/;
		if (!file.type.match(imageType)) return false;
		var img = document.createElement("img");
		img.file = file;
		opts.imageCreated(img);
		var reader = new FileReader();
		reader.onload = function(e) {
			img.src = e.target.result;
			img.onload = function() {
				img.title = ''+img.naturalWidth+'x'+img.naturalHeight+', '+prettySize(img.file.size)
			// $(img).attr('title',''+img.naturalWidth+'x'+img.naturalHeight+', '+prettySize(img.file.size));
			opts.imageLoaded(img);
			// log('loaded image')
			};
		};
		reader.readAsDataURL(file);
	}

	function dragEnter(e) {//log('dragEnter');
		clearTimeout(doc_leave_timer);
		e.preventDefault();
		opts.dragEnter(e);
	}

	function dragOver(e) {//log('dragOver');
		clearTimeout(doc_leave_timer);
		e.preventDefault();
		opts.docOver(e);
		opts.dragOver(e);
	}

	function dragLeave(e) {//log('dragLeave');
		clearTimeout(doc_leave_timer);
		opts.dragLeave(e);
		e.stopPropagation();
	}

	function docDrop(e) {//log('docDrop');
		e.preventDefault();
		opts.docLeave(e);
		return false;
	}

	function docEnter(e) {//log('docEnter');
		clearTimeout(doc_leave_timer);
		e.preventDefault();
		opts.docEnter(e);
		return false;
	}

	function docOver(e) {//log('docOver');
		clearTimeout(doc_leave_timer);
		e.preventDefault();
		opts.docOver(e);
		return false;
	}

	function docLeave(e) {//log('docLeave');
		doc_leave_timer = setTimeout(function(){
			opts.docLeave(e);
		}, 200);
	}

	function empty(){}

	// start upload statement

	function beforeEach(file) {
		return opts.beforeEach(file);
	}

	function afterAll() {
		return opts.afterAll();
	}

	function getIndexBySize(size) {
		for (var i = 0; i < files_count; i++) {
			if (files[i].size == size) {
				return i;
			}
		}

		return undefined;
	}

	function progress(e) { // xhr.upload event: progress
		if (e.lengthComputable) {
			var percentage = Math.round((e.loaded * 100) / e.total);

			if (this.currentProgress != percentage) {

				this.currentProgress = percentage;
				if (this.ctrl) {
					this.ctrl.update(this.currentProgress);
				}
				opts.progressUpdated(this.index, this.file, this.currentProgress);

				var elapsed = new Date().getTime();
				var diffTime = elapsed - this.currentStart;
				if (diffTime >= opts.refresh) {
					var diffData = e.loaded - this.startData;
					var speed = diffData / diffTime; // KB per second
					opts.speedUpdated(this.index, this.file, speed);
					this.startData = e.loaded;
					this.currentStart = elapsed;
				}
			}
		}
	}

	// Respond to an upload
	function upload() {
		if (arguments.length > 0 && $.isPlainObject(arguments[0])) {
			opts = $.extend( {}, default_opts, opts, arguments[0] );
			if (opts.files) {
				files = opts.files;
				files_count  = files.length;
			}
			if (opts.images) {
				images = opts.images;
			}
		}

		stop_loop = false;

		if (!files) {log(files);
			opts.error(errors[3]);
			return false;
		}

		var filesDone = 0,
			filesRejected = 0;

		if (files_count > opts.maxfiles && opts.queuefiles === 0) {
			opts.error(errors[1]);
			return false;
		}

		// Define queues to manage upload process
		var workQueue = [];
		var processingQueue = [];
		var doneQueue = [];

		// Add everything to the workQueue
		for (var i = 0; i < files_count; i++) {
			workQueue.push(i);
		}

		// Helper function to enable pause of processing to wait
		// for in process queue to complete
		var pause = function(timeout) {
				setTimeout(process, timeout);
				return;
		}

		// Process an upload, recursive
		var process = function() {

				var fileIndex;

				if (stop_loop) return false;

				// Check to see if are in queue mode
				if (opts.queuefiles > 0 && processingQueue.length >= opts.queuefiles) {

					return pause(opts.queuewait);

				} else {

					// Take first thing off work queue
					fileIndex = workQueue[0];
					workQueue.splice(0, 1);

					// Add to processing queue
					processingQueue.push(fileIndex);

				}

				try {
					if (beforeEach(files[fileIndex]) != false) {
						if (fileIndex === files_count) return;
						var reader = new FileReader(),
							max_file_size = 1024 * opts.maxfilesize;

						reader.index = fileIndex;
						if (files[fileIndex].size > max_file_size) {
							opts.error(errors[2], files[fileIndex], fileIndex);
							// Remove from queue
							processingQueue.forEach(function(value, key) {
								if (value === fileIndex) processingQueue.splice(key, 1);
							});
							filesRejected++;
							return true;
						}
						if (images.length == files_count) {
							reader.img = images[fileIndex];
						}
						reader.onloadend = send;
						reader.readAsBinaryString(files[fileIndex]);

					} else {
						filesRejected++;
					}
				} catch (err) {
					log(err);
					// Remove from queue
					processingQueue.forEach(function(value, key) {
						if (value === fileIndex) processingQueue.splice(key, 1);
					});
					opts.error(errors[0]);
					return false;
				}

				// If we still have work to do,
				if (workQueue.length > 0) {
					process();
				}

			};

		var send = function(e) { // FileReader event: loadend

			var fileIndex = ((typeof(e.srcElement) === "undefined") ? e.target : e.srcElement).index

			// Sometimes the index is not attached to the
			// event object. Find it by size. Hack for sure.
			if (e.target.index == undefined) {
				e.target.index = getIndexBySize(e.total);
			}

			if (e.target.img) {
				this.ctrl = createThrobber(e.target.img);
			}
			var self = this;

			var xhr = new XMLHttpRequest(),
					upload = xhr.upload,
					file = files[e.target.index],
					index = e.target.index,
					start_time = new Date().getTime(),
					formData = new FormData();

			if (opts.data) {
				$.each(opts.data, function(i, item) {
					formData.append(item.name, item.value);
				});
			}

			var name = (typeof file.id === "string" ? file.id : opts.field_name);
			if (typeof file.index !== "undefined") {
				formData.append(name + '_index', file.index)
			}
			if (typeof file.label === "string") {
				formData.append(name + '_label', file.label)
			}
			if (typeof file.tags === "string") {
				formData.append(name + '_tags', file.tags)
			}

			if (typeof file.lastModifiedDate === "object") {
				formData.append(name + '_ts', file.lastModifiedDate - 0)
			}

			formData.append(name, file)

			if (this.ctrl) upload.ctrl = this.ctrl;
			upload.index = index;
			upload.file = file;
			upload.downloadStartTime = start_time;
			upload.currentStart = start_time;
			upload.currentProgress = 0;
			upload.startData = 0;
			upload.addEventListener("progress", progress, false);

			xhr.open("POST", opts.url, true);

			// Add headers
			$.each(opts.headers, function(k, v) {
				xhr.setRequestHeader(k, v);
			});

			// xhr.sendAsBinary(binary);
			xhr.send(formData);

			opts.uploadStarted(index, file, files_count);

			xhr.onload = function() {
				if (xhr.responseText) {
					var now = new Date().getTime(),
							timeDiff = now - start_time,
							result = opts.uploadFinished(index, file, jQuery.parseJSON(xhr.responseText), timeDiff, xhr);
					filesDone++;
					//log('self:', typeof self, typeof this, typeof self.ctrl)
					if (self.ctrl) {
						self.ctrl.update(100);
						var canvas = self.ctrl.ctx.canvas;
						canvas.parentNode.removeChild(canvas);
					}

					// Remove from processing queue
					processingQueue.forEach(function(value, key) {
						if (value === fileIndex) processingQueue.splice(key, 1);
					});

					// Add to donequeue
					doneQueue.push(fileIndex);

					if (filesDone == files_count - filesRejected) {
						afterAll();
					}
					if (result === false) stop_loop = true;
				}
			};

		}

		// Initiate the processing loop
		process();

	}


	function createThrobber(img) {
		var offset = $(img).offset(), x = offset.left, y = offset.top;
		//var x = img.x;
		//var y = img.y;

		var canvas = document.createElement("canvas");
		img.parentNode.appendChild(canvas);
		canvas.width = img.width;
		canvas.height = img.height;
		var size = Math.min(canvas.height, canvas.width);
		canvas.style.top = y + "px";
		canvas.style.left = x + "px";
		canvas.classList.add("throbber");
		var ctx = canvas.getContext("2d");
		ctx.textBaseline = "middle";
		ctx.textAlign = "center";
		ctx.font = "15px monospace";
		ctx.shadowOffsetX = 0;
		ctx.shadowOffsetY = 0;
		ctx.shadowBlur = 14;
		ctx.shadowColor = "white";

		var ctrl = {};
		ctrl.ctx = ctx;
		ctrl.update = function(percentage) {
			var ctx = this.ctx;
			ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);
			ctx.fillStyle = "rgba(0, 0, 0, " + (0.8 - 0.8 * percentage / 100)+ ")";
			ctx.fillRect(0, 0, ctx.canvas.width, ctx.canvas.height);
			ctx.beginPath();
			ctx.arc(ctx.canvas.width / 2, ctx.canvas.height / 2,
					size / 6, 0, Math.PI * 2, false);
			ctx.strokeStyle = "rgba(255, 255, 255, 1)";
			ctx.lineWidth = size / 10 + 4;
			ctx.stroke();
			ctx.beginPath();
			ctx.arc(ctx.canvas.width / 2, ctx.canvas.height / 2,
					size / 6, -Math.PI / 2, (Math.PI * 2) * (percentage / 100) + -Math.PI / 2, false);
			ctx.strokeStyle = "rgba(0, 0, 0, 1)";
			ctx.lineWidth = size / 10;
			ctx.stroke();
			ctx.fillStyle = "white";
			ctx.baseLine = "middle";
			ctx.textAlign = "center";
			ctx.font = "10px monospace";
			ctx.fillText(percentage + "%", ctx.canvas.width / 2, ctx.canvas.height / 2);
		}
		ctrl.update(0);
		return ctrl;
	}


	// image upload
	$.extend({
		prettySize: prettySize,
		imgpreview: preview,
		imgupload: upload
	});

})(jQuery);
