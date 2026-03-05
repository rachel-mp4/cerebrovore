<script lang="ts">
  import type { Message } from "../types";
  import * as linkify from "linkifyjs";
  import diff from "fast-diff";
  interface Props {
    message: Message;
    mylocaltext?: string;
  }
  let { message, mylocaltext }: Props = $props();
  const escapeHTML = (text: string): string => {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
  };
  const convertLinksToMessageFrags = (body: string) => {
    const ebody = escapeHTML(body);
    const links = linkify.find(body, "url");
    const ll = links.length;
    if (ll === 0) {
      return [{ text: ebody, isLink: false, href: "", key: 0 }];
    }
    let res = [];
    let idx = 0;
    links.forEach((link) => {
      if (link.start > idx) {
        const beforeText = body.substring(idx, link.start);
        res.push({
          text: escapeHTML(beforeText),
          isLink: false,
          href: "",
          key: res.length,
        });
      }
      res.push({
        text: link.value,
        href: link.href,
        isLink: true,
        key: res.length,
      });
      idx = link.end;
    });
    if (idx < body.length) {
      const afterText = body.substring(idx);
      res.push({
        text: escapeHTML(afterText),
        isLink: false,
        key: res.length,
        href: "",
      });
    }
    return res;
  };
  let mfrags = $derived(convertLinksToMessageFrags(message.lrcdata.body));
  let diffs = $derived(
    message.lrcdata.pub && message.lrcdata.mine && mylocaltext
      ? diff(message.lrcdata.body, mylocaltext)
      : null,
  );
</script>

{#if diffs}
  {#each diffs as diff}
    {#if diff[0] === -1}
      <span class="removed">{diff[1]}</span>
    {:else if diff[0] === 0}
      <span>{diff[1]}</span>
    {:else}
      <span class="appended">{diff[1]}</span>
    {/if}
  {/each}
{:else}
  {#each mfrags as part (part.key)}
    {#if part.isLink}
      <a href={part.href} target="_blank" rel="noopener">{part.text}</a>
    {:else}
      {@html part.text}
    {/if}
  {/each}
{/if}
