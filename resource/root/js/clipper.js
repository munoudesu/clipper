var app = new Vue({
	el:"#app",
	data:{
		settings: {
		},
		defaultDuration: 120,
		channelId: "",
		clipRecommenders: "",
		clipVideoTitle: "",
		originUrl: "",
		clipsEtag: "",
		clips: null,
		clipsIndex: -1,
		youtubePlayer: null,
		lastYoutubePlayerStatus: -1,
		showSettingContent: false,
		showDescriptionContent: false,
		lastClip: null,
		random: false,
		randomIndexesIndex: -1,
		randomIndexes: null
	},
	mounted: function() {
		this.channelId = document.getElementById('channelId').value;
		let pagePropUrl = "../cache/" + this.channelId + ".json";
		axios.get(pagePropUrl).then(res => {
			this.clipsEtag = res.headers.etag;
			this.clips = res.data;
			this.loadSetting();
			this.youtubePlayerInit();
		}).catch(res => {
			console.log("can not get page property");
		});
	},
	methods: {
		initSetting: function() {
			this.settings[this.channelId] = {
				defaultDuration: this.defaultDuration
			};
		},
		loadSetting: function() {
			if (localStorage.getItem('settings')) {
				try {
					this.settings = JSON.parse(localStorage.getItem('settings'));
					if (this.settings[this.channelId])  {
						this.defaultDuration = this.settings[this.channelId].defaultDuration;
					} else {
						this.initSetting();
					}
				} catch(e) {
					localStorage.removeItem('settings');
					this.initSetting();
				}
			} else {
				this.initSetting();
			}
		},
		saveSetting: function() {
			if (this.settings[this.channelId]) {
				this.settings[this.channelId].defaultDuration = this.defaultDuration;
			} else {
				this.settings[this.channelId] = {
					defaultDuration: this.defaultDuration
				}
			}
			localStorage.setItem('settings', JSON.stringify(this.settings));
		},
		openSetting: function() {
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
		fixClipDuration: function(clip) {
			clip.Merge = 0;
			if (clip.End == 0) {
				clip.End = clip.Start + this.settings[this.channelId].defaultDuration;
			}
			while (true) {
				// 次のクリップの開始時間をチェック
				if (this.clipsIndex >= this.clips.length - 1) {
					// 次のクリップはもうない
					clip.Duration = clip.End - clip.Start;
					return;
				}
				let nextClip = JSON.parse(JSON.stringify(this.clips[this.clipsIndex + 1]));
				if (clip.VideoId != nextClip.VideoId) {
					// 次のクリップとビデオが一致しない
					clip.Duration = clip.End - clip.Start;
					return;
				}
				if (clip.End < nextClip.Start) {
					// 次のクリップの開始と被らない
					clip.Duration = clip.End - clip.Start;
					return;
				}
				// 次のクリップの開始と被る場合、次のクリップを含める
				if (nextClip.End == 0) {
					clip.End = nextClip.Start + this.settings[this.channelId].defaultDuration;
				} else {
					clip.End = nextClip.End;
				}
				clip.Recommenders.concat(nextClip.Recommenders);
				clip.Recommenders = Array.from(new Set(clip.Recommenders));
				clip.Merge += 1;
				// クリップのインデックスを進める
				this.incrementIndex();
			}
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
			if (this.lastClip != null) {
				this.clipsIndex -= this.lastClip.Merge;
			}
			this.clipsIndex -= 1;
			if (this.clipsIndex < 0) {
				this.clipsIndex = this.clips.length -1;
			}
		},
		incrementRandomIndexesIndex: function() {
			this.randomIndexesIndex += 1;
			if (this.randomIndexesIndex >= this.randomIndexes.length) {
				this.randomIndexesIndex = 0;
			}
		},
		decrementRandomIndexesIndex: function() {
			this.randomIndexesIndex -= 1;
			if (this.randomIndexesIndex < 0) {
				this.randomIndexesIndex = this.randomIndexes.length -1;
			}
		},
		getClip: function() {
			let clip = JSON.parse(JSON.stringify(this.clips[this.clipsIndex]));
			this.fixClipDuration(clip);
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
		getNextRandomClip: function() {
			this.incrementRandomIndexesIndex();
			this.clipsIndex = this.randomIndexes[this.randomIndexesIndex]
			return this.getClip();
		},
		getPreviousRandomClip: function() {
			this.decrementRandomIndexesIndex();
			this.clipsIndex = this.randomIndexes[this.randomIndexesIndex]
			return this.getClip();
		},
		updateClipView: function(clip) {
			this.originUrl = this.makeOriginUrl(clip);
			this.clipRecommenders = "推薦者:" + clip.Recommenders.join("さん, ") + "さん";
			this.clipVideoTitle = clip.Title;

		},
		randomEnable: function() {
			this.random = true;
			this.createRandomIndexes();
			this.randomIndexesIndex = 0;
			this.clipsIndex = this.randomIndexes[this.randomIndexesIndex];
			let clip = this.getClip();
			this.youtubePlayerLoad(clip);
		},
		randomDisable: function() {
			this.random = false;
			this.clipsIndex = 0;
			let clip = this.getClip();
			this.youtubePlayerLoad(clip);
		},
		createRandomIndexes: function() {
			this.randomIndexes = [];
			this.clipsIndex = 0;
			this.randomIndexes.push(this.clipsIndex);
			while (true) {
				this.getNextClip();
				this.randomIndexes.push(this.clipsIndex);
				if (this.clipsIndex >= this.clips.length - 1 || this.clipsIndex == 0) {
					break;
				}
			}
			//shuffle
			for (let i = this.randomIndexes.length - 1; i >= 0; i--){
				let rand = Math.floor(Math.random() * (i + 1));
				let tmp = this.randomIndexes[i];
				this.randomIndexes[i] = this.randomIndexes[rand];
				this.randomIndexes[rand] = tmp;
			}
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
				host: 'https://www.youtube.com',
				width: 854,
				height: 480,
				playerVars: {
					'autoplay': 1,
					'controls': 1,
					'start': clip.Start,
					'end': clip.End,
					'enablejsapi ': 1,
					'origin': location.protocol + '//' + location.hostname + "/"
				},
				events: {
					'onReady': this.onYoutubePlayerReady,
					'onStateChange': this.onYoutubePlayerStateChange,
					'onPlaybackQualityChange': this.onYoutubePlayerPlaybackQualityChange,
					'onError': this.onYoutubePlayerError
				}
			});
		},
		youtubePlayerLoadNext: function() {
			if (this.random == true) {
				let clip = this.getNextRandomClip();
				this.youtubePlayerLoad(clip);
			} else {
				let clip = this.getNextClip();
				this.youtubePlayerLoad(clip);
			}
		},
		youtubePlayerLoadPrevious: function() {
			if (this.random == true) {
				let clip = this.getPreviousRandomClip();
				this.youtubePlayerLoad(clip);
			} else {
				let clip = this.getPreviousClip();
				this.youtubePlayerLoad(clip);
			}
		},
		youtubePlayerLoad: function(clip) {
			this.lastClip = clip;
			this.updateClipView(clip)
			this.youtubePlayer.loadVideoById({
				videoId: clip.VideoId,
				startSeconds: clip.Start,
				endSeconds: clip.End
			});
		},
		onYoutubePlayerReady: function(event) {
			event.target.playVideo();
		},
		onYoutubePlayerStateChange: function(event) {
			let st = event.target.getPlayerState();
			if (this.lastYoutubePlayerStatus == YT.PlayerState.PLAYING && st == YT.PlayerState.ENDED) {
				this.youtubePlayerLoadNext();
			} else if (st == YT.PlayerState.ENDED) {
				let clip = this.getClip();
				if (this.youtubePlayer.getDuration() >= clip.Start) {
					this.youtubePlayerLoadNext();
				}
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
