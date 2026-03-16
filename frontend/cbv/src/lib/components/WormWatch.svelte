<script lang="ts">
  import type { WormWatchEntry } from "../types";
  import { onMount } from "svelte";
  import { b36encodenumber, nSecondsOutOfMax, nSecondsToHMS } from "../utils";
  import { WormWatchContext } from "../wormwatchcontext.svelte";
  import YoutubePlayer from "youtube-player";
  import type { YouTubePlayer as Player } from "youtube-player/dist/types";
  import {
    getVolume,
    getWormWatchVolume,
    onVolumeChange,
    onVolumeWormWatchChange,
  } from "../volume";
  import VolumeSettings from "./VolumeSettings.svelte";
  interface Props {
    ctx: WormWatchContext;
  }
  let { ctx }: Props = $props();

  let time: string | undefined = $state();
  let maxTime: string | undefined = $state();
  let interval: number | undefined;
  let volume: number = getVolume();
  let wwvolume: number = getWormWatchVolume();
  let showVolumeSettings = $state(false);

  onMount(() => {
    const startHandler = () => {
      interval = setInterval(() => {
        if (ctx.pause !== undefined) {
          return;
        }
        const nTime = ctx.getTimeToStart();
        if (nTime !== undefined) {
          const n = Math.floor(nTime / -1000);
          if (maxTime === undefined) {
            time = nSecondsToHMS(n);
          } else {
            time = nSecondsOutOfMax(n, maxTime);
          }
        } else {
          time = undefined;
        }
      }, 1000);
      const entry = ctx.wwqueue[ctx.playingIndex!];
      if (entry) {
        // duration from go is given as Nanosecond
        maxTime = nSecondsToHMS(Math.floor(entry.data.duration / 1000000000));
      }
      readyPlayerForAction(entry);
    };
    const pauseHandler = () => {
      player.pauseVideo();
    };
    const clearHandler = () => {
      clearInterval(interval);
      time = undefined;
      maxTime = undefined;
      interval = undefined;

      destroyPlayer();
    };
    ctx.addEventListener("start", startHandler);
    ctx.addEventListener("pause", pauseHandler);
    ctx.addEventListener("clear", clearHandler);
    const removeVC = onVolumeChange((e) => {
      volume = e.detail.volume;
      if (playerReady) {
        player.setVolume(Math.floor(volume * wwvolume * 100));
      }
    });
    const removeWWVC = onVolumeWormWatchChange((e) => {
      wwvolume = e.detail.volume;
      if (playerReady) {
        player.setVolume(Math.floor(volume * wwvolume * 100));
      }
    });
    return () => {
      ctx.removeEventListener("start", startHandler);
      ctx.removeEventListener("pause", pauseHandler);
      ctx.removeEventListener("clear", clearHandler);
      removeVC();
      removeWWVC();
    };
  });

  var player: Player;
  let playerReady = $state(false);
  let playerHeight = $state(0);
  let curID = $state();
  const readyPlayerForAction = (entry: WormWatchEntry) => {
    let reqsite = entry.data.site;
    switch (reqsite) {
      // youtube
      case 0: {
        if (!playerReady) {
          player = YoutubePlayer("worm-watch");
        }
        if (curID === entry.data.id) {
          onPlayerReady();
          return;
        }
        curID = entry.data.id;
        player.loadVideoById(entry.data.id).then(() => {
          player
            .setSize(entry.data.width ?? 576, entry.data.height ?? 324)
            .then(() => {
              player.setVolume(Math.floor(volume * wwvolume * 100)).then(() => {
                playerReady = true;
                playerHeight = entry.data.height ?? 324;
                onPlayerReady();
              });
            });
        });
        break;
      }
    }
  };
  const destroyPlayer = () => {
    player.destroy();
    playerReady = false;
    playerHeight = 0;
  };

  const onPlayerReady = () => {
    const st = ctx.getTimeToStart();
    if (st === undefined) {
      return;
    } else if (st < 0) {
      player.seekTo(st / -1000, true).then(player.playVideo);
    } else {
      setTimeout(() => {
        player.playVideo();
      }, st);
    }
  };
</script>

<div id="worm-watch"></div>

<div class="spacer" style:height="{playerHeight}px"></div>

<div class="sidebar-group">
  <div>
    <button
      onclick={() => {
        showVolumeSettings = !showVolumeSettings;
      }}>{showVolumeSettings ? "hide" : "show"} volume settings</button
    >
    {#if showVolumeSettings}
      <VolumeSettings />
    {/if}
  </div>
  {#if playerReady}
    <div class="sync-time">
      <button onclick={onPlayerReady}>sync up</button>
      <span>
        {time}{time !== undefined && maxTime !== undefined
          ? " / "
          : ""}{maxTime}
      </span>
    </div>
  {/if}
  {#if ctx.wwqueue.length !== 0 && !ctx.start}
    <div>paused</div>
  {/if}
  <ol class="queue">
    {#each ctx.wwqueue as entry, i}
      <li class="entry{i === ctx.playingIndex ? ' current' : ''}">
        {b36encodenumber(i)} -
        <a
          target="_blank"
          rel="noopener noreferrer"
          href="https://youtu.be/{entry.data.id}">{entry.data.title}</a
        >
      </li>
    {/each}
  </ol>
</div>
