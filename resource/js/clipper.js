var app = new Vue({
	el:"#app",
	data:{
		settings: {
			defaultDuration: 90
		},
		channelId: "",
		clipRecommenders: "",
		clipVideoTitle: "",
		originUrl: "",
		clips: null,
		clipsIndex: -1,
		youtubePlayer: null,
		lastYoutubePlayerStatus: -1,
		showSettingContent: false,
		showDescriptionContent: false
	},
	mounted: function() {
		this.channelId = document.getElementById('channelId').value;
		let pagePropUrl = "json/" + this.channelId + ".json";
		axios.get(pagePropUrl).then(res => {
			this.clips = res.data;
			this.loadSetting();
			this.youtubePlayerInit();
		}).catch(res => {
			console.log("can not get page property");
		});
	},
	methods: {
		loadSetting: function() {
			if (localStorage.getItem('settings')) {
				try {
					this.settings = JSON.parse(localStorage.getItem('settings'));
				} catch(e) {
					localStorage.removeItem('settings');
				}
			} 
		},
		saveSetting: function() {
			localStorage.setItem('settings', JSON.stringify(this.settings));
		},
		openSetting: function() {
			this.loadSetting();
			this.showSettingContent = true;
		},
		closeSetting: function() {
			this.showSettingContent = false;
			this.saveSetting();
		},
		openDescription: function() {
			this.showDescriptionContent = true;
		},
		closeDescription: function() {
			this.showDescriptionContent = false;
		},
		fixClipDuration: function(clip, nextClip) {
			if (clip.End == 0) {
				clip.End = clip.Start + this.settings.defaultDuration;
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
		incrementIndex: function() {
			this.clipsIndex += 1;
			if (this.clipsIndex >= this.clips.length) {
				this.clipsIndex = 0;
			}
		},
		decrementIndex: function() {
			this.clipsIndex -= 1;
			if (this.clipsIndex < 0) {
				this.clipsIndex = this.clips.length -1;
			}
		},
		getClip: function() {
			let clip = JSON.parse(JSON.stringify(this.clips[this.clipsIndex]));
			let nextClip = null;
			if (this.clipsIndex < this.clips.length - 1) {
				nextClip = JSON.parse(JSON.stringify(this.clips[this.clipsIndex + 1]));
			}
			this.fixClipDuration(clip, nextClip);
			console.log(clip);
			return clip;
		},
		getNextClip: function() {
			this.incrementIndex();
			return this.getClip();
		},
		getPreviousClip: function() {
			this.decrementIndex();
			return this.getClip();
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
		youtubePlayerLoadPrevious: function() {
			clip = this.getPreviousClip();
			this.updateClipView(clip)
			this.youtubePlayer.loadVideoById({
				videoId: clip.VideoId,
				startSeconds: clip.Start,
				endSeconds: clip.End
			});
		},
		onYoutubePlayerReady: function(event) {

		},
		onYoutubePlayerStateChange: function(event) {
			let st = event.target.getPlayerState();
			if (this.lastYoutubePlayerStatus == YT.PlayerState.PLAYING && st == YT.PlayerState.ENDED) {
				this.youtubePlayerLoadNext();
			}
			this.lastYoutubePlayerStatus = event.target.getPlayerState();
		},
		onYoutubePlayerPlaybackQualityChange: function(event) {

		},
		onYoutubePlayerError: function(event) {
			let error = event.data;
			if (error >= 100) {
				this.youtubePlayerLoadNext();
			}
		},
		playPreviousClip: function() {
			this.youtubePlayer.stopVideo();
			this.youtubePlayerLoadPrevious()
		},
		playNextClip: function() {
			this.youtubePlayer.stopVideo();
			this.youtubePlayerLoadNext()
		}
	}
});

function onYouTubeIframeAPIReady() {
	app.youtubePlayerCreate();
}
