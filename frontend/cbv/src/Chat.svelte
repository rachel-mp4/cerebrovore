<script lang="ts">
  import Receiver from "./lib/components/Receiver.svelte";
  import Transmitter from "./lib/components/Transmitter.svelte";
  import Console from "./lib/components/Console.svelte";
  import { WSContext } from "./lib/wscontext.svelte";

  interface Props {
    ismoderator: boolean;
    defaultnick: string | null;
    defaultcolor: number | null;
  }
  let { ismoderator, defaultnick, defaultcolor }: Props = $props();
  const url = window.location.href;
  // i think this should work for both http->ws and https->wss schemes, that's
  // why the magic number 4 is there
  const sansproto = url.slice(4);
  // i split on hashtag so that way user can load a page with a hashtag in it
  // (meaning jump to id)
  // and have the /ws not be interpreted as part of the hashtag
  const address = `ws${sansproto.split("#")[0]}`;
  const nick = localStorage.getItem("nick");
  const color = localStorage.getItem("color");
  const ctx = new WSContext(
    nick ?? defaultnick ?? "wanderer",
    color ? parseInt(color, 10) : (defaultcolor ?? 4534186),
  );
  ctx.connect(address);
</script>

{#if ctx.connected}
  <Receiver
    items={ctx.items}
    mylocaltext={ctx.curMsg}
    mylocalimage={ctx.curImageBlobURL}
    onmute={ctx.mute}
    onunmute={ctx.unmute}
    {ismoderator}
  />
  <Transmitter {ctx} />
  <Console log={ctx.log} />
{/if}
