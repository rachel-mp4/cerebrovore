<script lang="ts">
  import {
    getVolume,
    emitVolumeChange,
    getWormWatchVolume,
    emitVolumeWormWatchChange,
    getPingVolume,
    getWatcherVolume,
    emitVolumeWatcherChange,
    getFocusPingVolume,
    emitVolumePingChange,
    emitVolumeFocusPingChange,
  } from "../volume";
  let volume: number = $state(getVolume());
  let wormwatchvolume: number = $state(getWormWatchVolume());
  let watchervolume: number = $state(getWatcherVolume());
  let pingvolume: number = $state(getPingVolume());
  let focuspingvolume: number = $state(getFocusPingVolume());
  $effect(() => {
    emitVolumeChange(volume);
  });
  $effect(() => {
    emitVolumeWormWatchChange(wormwatchvolume);
  });
  $effect(() => {
    emitVolumeWatcherChange(watchervolume);
  });
  $effect(() => {
    emitVolumePingChange(pingvolume);
  });
  $effect(() => {
    emitVolumeFocusPingChange(focuspingvolume);
  });
</script>

<div class="volume-settings">
  <div class="volume-setting">
    <label for="global-volume">global volume</label>
    <input
      type="range"
      bind:value={volume}
      onchange={() => {
        localStorage.setItem("volume", String(volume));
      }}
      min="0"
      max="1"
      step="0.01"
      name="global-volume"
      id="global-volume"
    />
  </div>
  <div class="volume-setting">
    <label for="wormwatch-volume">wormwatch volume</label>
    <input
      type="range"
      bind:value={wormwatchvolume}
      onchange={() => {
        localStorage.setItem("volumewormwatch", String(wormwatchvolume));
      }}
      min="0"
      max="1"
      step="0.01"
      name="wormwatch-volume"
      id="wormwatch-volume"
    />
  </div>
  <div class="volume-setting">
    <label for="threadwatch-volume">threadwatch volume</label>
    <input
      type="range"
      bind:value={watchervolume}
      onchange={() => {
        localStorage.setItem("volumewatcher", String(watchervolume));
      }}
      min="0"
      max="1"
      step="0.01"
      name="threadwatch-volume"
      id="threadwatch-volume"
    />
  </div>
  <div class="volume-setting">
    <label for="ping-volume">ping volume</label>
    <input
      type="range"
      bind:value={pingvolume}
      onchange={() => {
        localStorage.setItem("volumeping", String(pingvolume));
      }}
      min="0"
      max="1"
      step="0.01"
      name="ping-volume"
      id="ping-volume"
    />
  </div>
  <div class="volume-setting">
    <label for="focusping-volume">focused ping volume</label>
    <input
      type="range"
      bind:value={focuspingvolume}
      onchange={() => {
        localStorage.setItem("volumefocusping", String(focuspingvolume));
      }}
      min="0"
      max="1"
      step="0.01"
      name="focusping-volume"
      id="focusping-volume"
    />
  </div>
</div>
