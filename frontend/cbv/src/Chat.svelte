<script lang="ts">
  import { onDestroy } from "svelte";
  import Receiver from "./lib/components/Receiver.svelte";
  import Transmitter from "./lib/components/Transmitter.svelte";
  import { WSContext } from "./lib/wscontext.svelte";
  let url = $state(window.location.href);
  // i think this should work for both http->ws and https->wss schemes, that's
  // why the magic number 4 is there
  let sansproto = url.slice(4);
  // i split on hashtag so that way user can load a page with a hashtag in it
  // (meaning jump to id)
  // and have the /ws not be interpreted as part of the hashtag
  let address = $derived(`ws${sansproto.split("#")[0]}/ws`);
  let ctx = $derived(new WSContext("wanderer", 4534186));

  $effect(() => {
    ctx?.disconnect();
    if (address) {
      ctx.connect(address);
    }
  });
  // not sure why we'd ever destroy since not SPA but might as well throw it in
  onDestroy(() => ctx?.disconnect());
</script>

{#if ctx}
  <Receiver
    items={ctx.items}
    mylocaltext={ctx.curMsg}
    onmute={ctx.mute}
    onunmute={ctx.unmute}
  />
  <Transmitter {ctx} />
{/if}
