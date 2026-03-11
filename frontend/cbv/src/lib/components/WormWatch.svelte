<script lang="ts">
  import type { WormWatchEntry } from "../types";
  import { onMount } from "svelte";
  import { b36encodenumber } from "../utils";
  import { WormWatchContext } from "../wormwatchcontext.svelte";
  import YoutubePlayer from "youtube-player";
  import type { YouTubePlayer as Player } from "youtube-player/dist/types";
  interface Props {
    ctx: WormWatchContext;
  }
  let { ctx }: Props = $props();
  onMount(() => {
    const startHandler = () => {
      const entry = ctx.wwqueue[ctx.playingIndex!];
      readyPlayerForAction(entry);
    };
    const pauseHandler = () => {
      player.pauseVideo();
    };
    ctx.addEventListener("start", startHandler);
    ctx.addEventListener("pause", pauseHandler);
    ctx.addEventListener("clear", destroyPlayer);
  });

  var player: Player;
  let playerReady = false;
  let playerHeight = $state(0);
  const readyPlayerForAction = (entry: WormWatchEntry) => {
    let reqsite = entry.data.site;
    switch (reqsite) {
      // youtube
      case 0: {
        if (!playerReady) {
          player = YoutubePlayer("worm-watch");
        }
        player.loadVideoById(entry.data.id).then(() => {
          player
            .setSize(entry.data.width ?? 576, entry.data.height ?? 324)
            .then(() => {
              playerReady = true;
              playerHeight = entry.data.height ?? 324;
              onPlayerReady();
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
  <button onclick={onPlayerReady}>sync up</button>
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
