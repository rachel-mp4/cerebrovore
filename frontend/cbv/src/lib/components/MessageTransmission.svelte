<script lang="ts">
  import type { Message } from "../types";
  import { parse, render } from "../liveparse";
  interface Props {
    message: Message;
    mylocaltext?: string;
  }
  let { message, mylocaltext }: Props = $props();

  // let diffs = $derived(
  //   !message.lrcdata.pub && message.lrcdata.mine && mylocaltext
  //     ? diff(message.lrcdata.body, mylocaltext)
  //     : null,
  // );

  const div = document.createElement("div");
  const zizzify = (s: string): string => {
    if (s.length < 2222) {
      return render(parse(s));
    } else {
      div.textContent = s;
      return div.innerHTML;
    }
  };
</script>

<div>
  <!-- {#if diffs} -->
  <!--   {#each diffs as diff} -->
  <!--     {#if diff[0] === -1} -->
  <!--       <span class="removed">{diff[1]}</span> -->
  <!--     {:else if diff[0] === 0} -->
  <!--       <span>{diff[1]}</span> -->
  <!--     {:else} -->
  <!--       <span class="appended">{diff[1]}</span> -->
  <!--     {/if} -->
  <!--   {/each} -->
  {#if message.renderedHTML}
    {@html message.renderedHTML}
  {:else if mylocaltext}
    {@html zizzify(mylocaltext)}
  {:else}
    {@html zizzify(message.lrcdata.body)}
  {/if}
</div>
