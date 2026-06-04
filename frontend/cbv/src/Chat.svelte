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
  let imageURL: string | undefined = $state();
  const convertFileToImageItem = (blob: File) => {
    cancelimagepost();
    const blobUrl = URL.createObjectURL(blob);
    ctx.initImage(blob, blobUrl);
    imageURL = blobUrl;
  };
  const cancelimagepost = () => {
    if (imageURL) {
      URL.revokeObjectURL(imageURL);
    }
    ctx.cancelImage();
    imageURL = undefined;
  };
  const uploadimage = (alt: string | undefined) => {
    ctx.pubImage(alt);
    if (imageURL) {
      URL.revokeObjectURL(imageURL);
    }
    imageURL = undefined;
  };
  const pastifier = (event: ClipboardEvent) => {
    const items = event.clipboardData?.items;
    if (items === undefined) {
      return;
    }
    for (const item of items) {
      if (item.type.startsWith("image/")) {
        const blob = item.getAsFile();
        if (blob === null) {
          return;
        }
        event.preventDefault();
        convertFileToImageItem(blob);
      }
    }
  };
  document.addEventListener("paste", pastifier);
</script>

{#if ctx.connected}
  <Receiver
    items={ctx.items}
    mylocalid={ctx.curMsgId}
    mylocaltext={ctx.curMsg}
    mylocalimage={ctx.curImageBlobURL}
    onmute={ctx.mute}
    onunmute={ctx.unmute}
    {ismoderator}
    {cancelimagepost}
    {uploadimage}
    {ctx}
  />
  <Transmitter {ctx} {defaultnick} {defaultcolor} {convertFileToImageItem} />
  <Console log={ctx.log} />
{/if}
