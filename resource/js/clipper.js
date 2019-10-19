var app = new Vue({
	el:"#app",
	data:{
		defaultDuration: 90,
		channelId: "",
		clipRecommenders: "",
		clipVideoTitle: "",
		originUrl: "",
		clips: null,
		clipsIndex: 0,
		youtubePlayer: null,
		lastYoutubePlayerStatus: -1
	},
	mounted: function() {
		this.channelId = document.getElementById('channelId').value;
		let pagePropUrl = "json/" + this.channelId + ".json";
		axios.get(pagePropUrl).then(res => {
			this.clips = res.data;
			console.log("0000000000");
			this.youtubePlayerInit();
		}).catch(res => {
			console.log(res);
			console.log("can not get page property");
		});
	},
	methods: {
		fixClipDuration: function(clip, nextClip) {
			if (clip.End == 0) {
				clip.End = clip.Start + this.defaultDuration;
			}
			if (nextClip != null && clip.VideoId == nextClip.VideoId) {
				if (clip.End > nextClip.Start) {
					clip.End = nextClip.Start;
				}
			}
			clip.Duration = clip.End - clip.Start;
		},
		makeOriginUrl: function(clip) {
			return "https://youtu.be/" + clip.VideoId + "?t=" + clip.Start;
		},
		getNextClip: function() {
			console.log("====");
			let clip = JSON.parse(JSON.stringify(this.clips[this.clipsIndex]));
			let nextClip = null;
			if (this.clipsIndex < this.clips.length - 1) {
				nextClip = JSON.parse(JSON.stringify(this.clips[this.clipsIndex + 1]));
			}
			console.log(clip);
			console.log(nextClip);
			this.fixClipDuration(clip, nextClip);
			this.clipsIndex += 1;
			if (this.clipsIndex == this.clips.length) {
				this.clipsIndex = 0;
			}
			console.log(clip);
			return clip;
		},
		updateClipView: function(clip) {
			this.originUrl = this.makeOriginUrl(clip);
			this.clipRecommenders = "推薦者:" + clip.Recommenders.join("さん, ") + "さん";
			this.clipVideoTitle = clip.Title;

		},
		youtubePlayerInit: function() {
			let tag = document.createElement('script');
			tag.src = "https://www.youtube.com/iframe_api";
			let firstScriptTag = document.getElementsByTagName('script')[0];
			firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);
		},
		youtubePlayerCreate: function() {
			console.log("11111111111");
			let clip = this.getNextClip();
			this.updateClipView(clip)
			this.youtubePlayer = new YT.Player('player', {
				videoId: clip.VideoId,
				playerVars: { 'autoplay': 1, 'controls': 1, 'start': clip.Start, 'end': clip.End },
				events: {
					'onReady': this.onYoutubePlayerReady,
					'onStateChange': this.onYoutubePlayerStateChange,
					'onPlaybackQualityChange': this.onYoutubePlayerPlaybackQualityChange,
					'onError': this.onYoutubePlayerError
				}
			});
		},
		youtubePlayerLoadNext: function() {
			clip = this.getNextClip();
			this.updateClipView(clip)
			this.youtubePlayer.loadVideoById({
				videoId: clip.VideoId,
				startSeconds: clip.Start,
				endSeconds: clip.End
			});
		},
		onYoutubePlayerReady: function(event) {
				console.log(event);

		},
		onYoutubePlayerStateChange: function(event) {
			console.log(event);
			let st = event.target.getPlayerState();
			if (this.lastYoutubePlayerStatus == YT.PlayerState.PLAYING && st == YT.PlayerState.ENDED) {
				console.log("xxxxxx");
				this.youtubePlayerLoadNext();
			}
			this.lastYoutubePlayerStatus = event.target.getPlayerState();
		},
		onYoutubePlayerPlaybackQualityChange: function(event) {
				console.log(event);

		},
		onYoutubePlayerError: function(event) {
			console.log(event);
			let error = event.data;
			if (error >= 100) {
				console.log("yyyyyyy");
				this.youtubePlayerLoadNext();
			}
		}
	}
});

function onYouTubeIframeAPIReady() {
	console.log("zzzzzzzzzzz");
	app.youtubePlayerCreate();
}
