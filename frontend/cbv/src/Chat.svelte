<script lang="ts">
  import { onDestroy } from "svelte";
  import Receiver from "./lib/components/Receiver.svelte";
  import Transmitter from "./lib/components/Transmitter.svelte";
  import { WSContext } from "./lib/wscontext.svelte";
  let url = $state(window.location.href);
  // i think this should work for both http->ws and https->wss schemes, that's
  // why the magic number 4 is there
  let address = $derived(`ws${url.slice(4)}/ws`);
  let ctx = $derived(
    new WSContext(
      "i don't think this matters",
      "cool.org",
      "wanderer",
      4534186,
    ),
  );

  $effect(() => {
    ctx?.disconnect();
    if (address) {
      ctx.connect(address);
    }
  });
  // not sure why we'd ever destroy since not SPA but might as well throw it in
  onDestroy(() => ctx?.disconnect());
  let innerWidth = $state(0);
  let isDesktop = $derived(innerWidth > 1000);
</script>

<svelte:window bind:innerWidth />
<main id="transceiver" class={isDesktop ? "desktop" : "mobile"}>
  {#if !ctx?.connected}
    <div>
      connecting... <span class="error-message"
        >probably something went wrong if you can read me, maybe refresh?</span
      >
    </div>
  {/if}
  {#if ctx}
    {#if ctx.items.length === 0 && ctx.connected}
      <div>connecting...</div>
      <div>and you're connected.</div>
    {/if}
    <Receiver
      items={ctx.items}
      mylocaltext={ctx.curMsg}
      onmute={ctx.mute}
      onunmute={ctx.unmute}
    />
    <Transmitter {ctx} />
  {/if}
</main>

<style>
  .error-message {
    opacity: 0;
    animation: fadeIn 0.1s ease-in-out 1s forwards;
  }
  @keyframes fadeIn {
    to {
      opacity: 1;
    }
  }

  #transceiver {
    position: relative;
    display: flex;
    flex-direction: column;
  }
  #transceiver.desktop {
    height: 100vh;
  }
  #transceiver.mobile {
    height: calc(100vh - 2.25rem);
    height: calc(100dvh - 2.25rem);
    overflow-y: scroll;
  }
</style>
