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
  let ui = $state(navigator.userActivation.hasBeenActive);
  if (!ui) {
    // @ts-expect-error this works in firefox
    if (navigator.getAutoplayPolicy("mediaelement") === "allowed") {
      ui = true;
    }
  }

  onMount(() => {
    const startHandler = () => {
      if (interval === undefined) {
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
      }
      const entry = ctx.wwqueue[ctx.playingIndex!];
      if (entry) {
        // duration from go is given as Nanosecond
        maxTime = nSecondsToHMS(Math.floor(entry.data.duration / 1000000000));
      }
      readyPlayerForAction(entry);
    };
    const pauseHandler = () => {
      if (ctx.pause === undefined) {
        console.error("pause shouldn't be undefined here");
      }
      const entry = ctx.wwqueue[ctx.playingIndex!];
      if (entry) {
        // duration from go is given as Nanosecond
        maxTime = nSecondsToHMS(Math.floor(entry.data.duration / 1000000000));
      }
      if (maxTime === undefined) {
        time = nSecondsToHMS((ctx.pause ?? 0) / 1000);
      } else {
        time = nSecondsOutOfMax((ctx.pause ?? 0) / 1000, maxTime);
      }
      readyPlayerForPause(entry);
    };
    const clearHandler = () => {
      clearInterval(interval);
      time = undefined;
      maxTime = undefined;
      interval = undefined;

      destroyPlayer();
    };
    const interact = () => {
      ui = true;
      if (playerReady) {
        player?.setVolume(ui ? Math.floor(volume * wwvolume * 100) : 0);
      }
      window.removeEventListener("pointerdown", interact);
    };
    const initplay = window.addEventListener("pointerdown", interact);
    if (ctx.isPlaying()) {
      startHandler();
    }
    ctx.addEventListener("start", startHandler);
    ctx.addEventListener("pause", pauseHandler);
    ctx.addEventListener("clear", clearHandler);
    const removeVC = onVolumeChange((e) => {
      volume = e.detail.volume;
      if (playerReady) {
        player?.setVolume(ui ? Math.floor(volume * wwvolume * 100) : 0);
      }
    });
    const removeWWVC = onVolumeWormWatchChange((e) => {
      wwvolume = e.detail.volume;
      if (playerReady) {
        player?.setVolume(ui ? Math.floor(volume * wwvolume * 100) : 0);
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

  var player: Player | undefined = $state();
  let playerReady = $state(false);
  let scale = $state(1);
  let playerHeight = $state(0);
  let playerAspect = $state(0);
  let pointerstarty: number | undefined = $state();
  let pointerstartx: number | undefined;
  let pointercurx: number | undefined;
  let pointercury: number | undefined;
  let curID: string | undefined;
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
        player?.loadVideoById(entry.data.id).then(() => {
          player
            ?.setSize(
              (entry.data.width ?? 576) * scale,
              (entry.data.height ?? 324) * scale,
            )
            .then(() => {
              player
                ?.setVolume(ui ? Math.floor(volume * wwvolume * 100) : 0)
                .then(() => {
                  playerReady = true;
                  playerHeight = entry.data.height ?? 324;
                  playerAspect = (entry.data.width ?? 576) / playerHeight;
                  onPlayerReady();
                });
            });
        });
        break;
      }
    }
  };
  const readyPlayerForPause = (entry: WormWatchEntry) => {
    let reqsite = entry.data.site;
    switch (reqsite) {
      // youtube
      case 0: {
        if (!playerReady) {
          player = YoutubePlayer("worm-watch");
        }
        if (curID === entry.data.id) {
          onPlayerPause();
          return;
        }
        curID = entry.data.id;
        player?.loadVideoById(entry.data.id).then(() => {
          console.log("loaded");
          player
            ?.setSize(
              (entry.data.width ?? 576) * scale,
              (entry.data.height ?? 324) * scale,
            )
            .then(() => {
              player
                ?.setVolume(Math.floor(volume * wwvolume * 100))
                .then(() => {
                  playerReady = true;
                  playerHeight = entry.data.height ?? 324;
                  playerAspect = (entry.data.width ?? 576) / playerHeight;
                  onPlayerPause();
                });
            });
        });
        break;
      }
    }
  };
  const destroyPlayer = () => {
    player?.destroy().then(() => {
      player = undefined;
    });
    playerReady = false;
    playerHeight = 0;
    playerAspect = 0;
  };

  const onPlayerReady = async () => {
    const st = ctx.getTimeToStart();
    if (st === undefined) {
      return;
    } else if (st < 0) {
      const listener = player?.on("stateChange", (event: any) => {
        console.log(event);
        if (event.data === 1) {
          const st = ctx.getTimeToStart();
          if (st === undefined) {
            console.error("i hate life");
            return;
          }
          console.log("beep");
          player?.seekTo(st / -1000, true);
          // @ts-expect-error this function exists, i think the typescript is no good
          player?.off(listener);
        }
      });
      player?.playVideo();
    } else {
      setTimeout(() => {
        player?.playVideo();
      }, st);
    }
  };

  const onPlayerPause = () => {
    const pause = ctx.pause;
    if (pause === undefined) {
      return;
    } else if (pause < 0) {
      player?.pauseVideo();
    } else {
      player?.pauseVideo().then(() => player?.seekTo(pause / 1000, true));
    }
  };
  let guesscale = $state(1);
</script>

<div id="worm-watch"></div>

<div
  id="worm-watch-resize"
  role="separator"
  aria-orientation="vertical"
  class="{pointerstarty !== undefined ? 'currently-resizing' : ''} {player &&
  !ui
    ? 'autoplay-disabled'
    : ''}"
  style="position:absolute; z-index:0; top: 0; right: 0; cursor: {pointerstarty !==
  undefined
    ? 'grabbing'
    : 'grab'};"
  style:height="{playerHeight * guesscale}px"
  style:width="{playerAspect * playerHeight * guesscale}px"
  onpointerdown={(event: PointerEvent) => {
    if (event.button !== 0) {
      return;
    }
    const spacer = event.target as HTMLDivElement;
    if (spacer) {
      spacer.setPointerCapture(event.pointerId);
      pointerstarty = event.clientY;
      pointerstartx = event.clientX;
      pointercurx = event.clientX;
      pointercury = event.clientY;
    }
  }}
  onpointermove={(event: PointerEvent) => {
    if (pointerstarty && pointerstartx) {
      pointercury = event.clientY;
      pointercurx = event.clientX;
      const sx = Math.min(
        Math.max(
          (scale * (window.innerWidth - pointercurx)) /
            (window.innerWidth - pointerstartx),
          0.5,
        ),
        2,
      );
      const sy = Math.min(
        Math.max((scale * pointercury) / pointerstarty, 0.5),
        2,
      );
      guesscale = Math.max(sx, sy);
    }
  }}
  onpointerup={(event: PointerEvent) => {
    if (pointerstartx && pointerstarty) {
      pointercury = event.clientY;
      pointercurx = event.clientX;
      const sx = Math.min(
        Math.max(
          (scale * (window.innerWidth - pointercurx)) /
            (window.innerWidth - pointerstartx),
          0.5,
        ),
        2,
      );
      const sy = Math.min(
        Math.max((scale * pointercury) / pointerstarty, 0.5),
        2,
      );
      scale = Math.max(sx, sy);
      player?.setSize(
        playerHeight * playerAspect * scale,
        playerHeight * scale,
      );
    }
    pointerstarty = undefined;
    pointercury = undefined;
    pointerstartx = undefined;
    pointercurx = undefined;
  }}
></div>

<div
  role="separator"
  aria-orientation="vertical"
  style:height="{playerHeight * scale}px"
></div>

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
