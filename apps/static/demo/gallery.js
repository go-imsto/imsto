
function format_size (size) {
	if (size >= 1073741824) {
		return (size / 1073741824).toFixed(2) + ' Gb';
	}
	if (size >= 1048576) {
		return (size / 1048576).toFixed(2) + ' Mb';
	}
	if (size >= 1024) {
		return (size / 1024).toFixed(2) + ' Kb';
	}
	return size + ' bytes';
};

var allow_delete = true, default_roof = 'demo', roof, page_size = 20;
var post_data = {roof:default_roof, sort_name:'created', sort_order:'desc', tags: ''};

$(document).ready(function(){
	var pageClick = function(page) {
		//arr = $("input:checked");
		//post_data.source = arr[0].value;
		//post_data.sort_order = arr[1].value;
		loadImages(page,post_data);
	};

	var searchClick = function() {
		post_data.tags = $('input[name=tags]').val();
		loadImages(1, post_data);
	};

	$('form').submit(function(e){
		searchClick();
		return false;
	});

	$('#tags').on("blur", function(e){
		searchClick();
	});

	$.getJSON('/imsto/roofs', function(res){
		// console.log(typeof res.roofs)
		var roofs = {};
		res.data.forEach(function(name, idx){roofs[name]=name.toUpperCase()})
		$('#radios1').radioButtons({
			data: roofs,
			name: 'roof',
			selected: default_roof
		});
		$('input[type=radio]', '#radios1').change(function(e){
			//console.log(this);
			loadImages(1, post_data);
		}).get(0).checked = true;
		loadImages(1, post_data);
	});

	function loadImages(page, post_data) {
		$("#rsp-status").html('loading...');
		roof = $('input[type=radio]:checked', '#radios1').val();
		$.get('/imsto/'+roof+'/metas?page='+page+'&rows='+page_size+'&sort_name='+post_data.sort_name+'&sort_order='+post_data.sort_order+'&tags='+post_data.tags, function(res){
			// log(res);
			if (typeof res.meta.error == "string") {
				console && console.log(res.error)
				$("#rsp-status").empty();
				$("#image-list").empty();
				return
			}
			if (!$.isArray(res.data) || res.data.length == 0) {
				console && console.log("items is empty");
				$("#image-list").empty().text("nothing found");
				$("#rsp-status").empty();
				return false;
			}
			var page_count = Math.round(res.meta.total/page_size); log(page_count);
			if (page_count > 1) {
				$("#pager").pager({ pagenumber: page, pagecount: page_count, buttonClickCallback: pageClick });
			} else {
				$("#pager").empty();
			}

			$("#rsp-status").empty();
			// var url_prefix = res.meta.url_prefix ? res.meta.url_prefix.replace('http:', location.protocol) : "/thumb/"; // test only, 正式环境需要修改
			var url_prefix = res.meta.stageHost ? '//'+res.meta.stageHost+'/show/' : '/show/'
			$("#image-list").empty();

			$.each(res.data, function(i, item){//console.log(item);
				var alt = (item.meta.width && item.meta.height) ? (item.meta.width+'x'+item.meta.height) : '', title = item.note ? item.note : '';

				if (item.created) {
					alt += 'date: ' + item.created;
				}
				var _li = $("<li></li>").attr("id", "li_iid" + item.id).appendTo("#image-list"),
					_a = $("<a></a>").attr("href", url_prefix + 'orig/' + item.path).attr("title", title+" ( "+alt+" )").attr('rel','box'),
					_img = $("<img/>").attr("src", url_prefix + 'w120/' + item.path).attr("srcset", url_prefix + 'w240/' + item.path + " 2x").attr('alt',alt).attr("class", ""),
					_txt = $("<span></span>").attr("class", "lbl").attr("title",item.size).text(format_size(item.size)),
					_btn = $("<span></span>").attr("class", "btn btn_delete ui-corner-all ui-icon ui-icon-trash").attr("title", "删除")
						.click(function(){
							if (!confirm("确认要删除此图片么，此操作将不可恢复？")) return false;
							// 删除指定的图片
							var id = $(this).parent().attr("id").substr(6);
							$.ajax({
								url: "/imsto/"+roof+"/"+id,
								headers: {"X-Access-Key": api_key},
								type: 'DELETE',
								cache: false,
								dataType: 'json',
								success: function(res) {
									//log(data, $("#li_iid" + id));
									if (res.meta.ok === true) {
										$("#li_iid" + id).fadeOut(function(){$(this).remove();});
									}
									else alertAjaxResult(res);
								}
							});

							return false;
						}).hide();
				_a.append(_img).appendTo(_li);//log(_img);
				_txt.appendTo(_li);
				if (allow_delete) {
					_li.append(_btn).hover(function(){
						_btn.fadeIn();
					},function(){
						_btn.fadeOut();
					});
				}

			});

			// colorbox
			$("#image-list a").colorbox();

		}, 'json');
	}

	$(".button").button();

});
