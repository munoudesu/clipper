var app = new Vue({
        el:"#app",
        data:{
                defaultDuration: 60,
                recommender: "",
                clipUrl: "",
		clips: nil,
		clipsIndex: 0
        },
        created: function() {
                this.clips = Clips;
        },
        mounted: function() {
                this.playClip();
        },
        methods: {
                fixClipDuration: function(clip) {
			if (clip.End == 0) {
				clip.End = clip.Start + this.defaultDuration;
			}
			clip.Duration = clip.End - clip.Start;
                },
                makeClipUrl: function(clip) {
			return "https://www.youtube.com/embed/" + clip.VideoId + "?start=" + clip.Start + "&end=" + clip.End + "autoplay=1";
                },
		playClip: function() {
			clip = JSON.parse(JSON.stringify(this.clips[this.clipsIndex]));
			this.fixClipDuration(clip);
			this.clipUrl = this.makeClipUrl(clip);
			this.clipRecommenders = clip.recommenders;
			this.clipVideoTitle = clip.VideoTitle;
			this.clipsIndex += 1;
			if this.clipsIndex == clips.length {
				this.clipsIndex = 0;
			}
			setTimeout(() => { this.playClip() }, clip.Duration * 1000);
		}
	}
}
