<script lang="ts">
  import type { Message } from "../types";
  import * as linkify from "linkifyjs";
  import "linkify-plugin-hashtag";
  import diff from "fast-diff";
  interface Props {
    message: Message;
    mylocaltext?: string;
  }
  let { message, mylocaltext }: Props = $props();

  // const div = document.createElement("div");
  // const escapeHTML = (text: string): string => {
  //   div.textContent = text;
  //   return div.innerHTML;
  // };

  // const badHashtag = new RegExp(/[^0-9A-Za-z]/);

  // const convertLinksToMessageFrags = (body: string) => {
  //   const ebody = escapeHTML(body);
  //   const links = linkify.find(body);
  //   const ll = links.length;
  //   if (ll === 0) {
  //     return [
  //       { text: ebody, isLink: false, isHashtag: false, href: "", key: 0 },
  //     ];
  //   }
  //   let res = [];
  //   let idx = 0;
  //   links.forEach((link) => {
  //     if (link.start > idx) {
  //       const beforeText = body.substring(idx, link.start);
  //       res.push({
  //         text: escapeHTML(beforeText),
  //         isLink: false,
  //         href: "",
  //         key: res.length,
  //       });
  //     }
  //
  //     switch (link.type) {
  //       case "email": {
  //         res.push({
  //           text: link.value,
  //           href: "",
  //           isLink: false,
  //           key: res.length,
  //         });
  //         break;
  //       }
  //
  //       case "url": {
  //         const v = link.value.startsWith("http")
  //           ? link.value
  //           : `https://${link.value}`;
  //         res.push({
  //           text: link.value,
  //           href: v,
  //           isLink: true,
  //           key: res.length,
  //         });
  //         break;
  //       }
  //
  //       case "hashtag": {
  //         const tag = link.value.slice(1);
  //         const badp = tag.search(badHashtag);
  //         switch (badp) {
  //           case -1: {
  //             res.push({
  //               text: link.value,
  //               href: `/p/${tag}`,
  //               isLink: true,
  //               key: res.length,
  //             });
  //             break;
  //           }
  //           case 0: {
  //             res.push({
  //               text: link.value,
  //               href: "",
  //               isLink: false,
  //               key: res.length,
  //             });
  //             break;
  //           }
  //           default: {
  //             const tags = tag.slice(0, badp);
  //             res.push({
  //               text: `#${tags}`,
  //               href: `/p/${tags}`,
  //               isLink: true,
  //               key: res.length,
  //             });
  //             const ntags = tag.slice(badp);
  //             res.push({
  //               text: `${ntags}`,
  //               href: "",
  //               isLink: false,
  //               key: res.length,
  //             });
  //             break;
  //           }
  //         }
  //       }
  //     }
  //     idx = link.end;
  //   });
  //   if (idx < body.length) {
  //     const afterText = body.substring(idx);
  //     res.push({
  //       text: escapeHTML(afterText),
  //       isLink: false,
  //       key: res.length,
  //       href: "",
  //     });
  //   }
  //   return res;
  // };
  // let mfrags = $derived(convertLinksToMessageFrags(message.lrcdata.body));
  let diffs = $derived(
    !message.lrcdata.pub && message.lrcdata.mine && mylocaltext
      ? diff(message.lrcdata.body, mylocaltext)
      : null,
  );
</script>

<div>
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
  {:else if message.renderedHTML}
    {@html message.renderedHTML}
    <!-- {#each mfrags as part (part.key)} -->
    <!-- {#if part.isLink} -->
    <!-- <a href={part.href} target="_blank" rel="noopener">{part.text}</a> -->
    <!-- {:else} -->
    <!-- {@html part.text} -->
    <!-- {/if} -->
    <!-- {/each} -->
  {:else}
    {message.lrcdata.body}
  {/if}
</div>
