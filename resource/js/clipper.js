//      var tag = document.createElement('script');
//      tag.src = "https://www.youtube.com/iframe_api";
//      var firstScriptTag = document.getElementsByTagName('script')[0];
//      firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);

var app = new Vue({
	el:"#app",
	data:{
		defaultDuration: 180,
		channelId: "",
		clipRecommenders: "",
		clipVideoTitle: "",
		clipUrl: "",
		clips: null,
		clipsIndex: 0
	},
	mounted: function() {
		this.channelId = document.getElementById('channelId').value;
		let pagePropUrl = "json/" + this.channelId + ".json"
		axios.get(pagePropUrl).then(res => {
			this.clips = res.data;
			this.playClip();
		}).catch(res => {
			console.log(res);
			console.log("can not get page property");
		});
	},
	methods: {
		fixClipDuration: function(clip) {
			if (clip.End == 0) {
				clip.End = clip.Start + this.defaultDuration;
			}
			clip.Duration = clip.End - clip.Start;
		},
		makeClipUrl: function(clip) {
			return "https://www.youtube.com/embed/" + clip.VideoId + "?start=" + clip.Start + "&end=" + clip.End + "&autoplay=1";
		},
		playClip: function() {
			clip = JSON.parse(JSON.stringify(this.clips[this.clipsIndex]));
			this.fixClipDuration(clip);
			this.clipUrl = this.makeClipUrl(clip);
			this.clipRecommenders = "推薦者:" + clip.Recommenders.join("さん, ") + "さん";
			this.clipVideoTitle = clip.VideoTitle;
			this.clipsIndex += 1;
			if (this.clipsIndex == this.clips.length) {
				this.clipsIndex = 0;
			}
			setTimeout(() => { this.playClip() }, clip.Duration * 1000);
		}
	}
});
