<script lang="ts">
  import VolumeSettings from "./lib/components/VolumeSettings.svelte";
  import { diffzizz } from "./lib/liveparse";
  const ntv = localStorage.getItem("new-threads");
  let wantsNewThreads = $state(ntv !== "");
  var nww = localStorage.getItem("nonewormwatch");
  if (nww === null) {
    nww = "";
    localStorage.setItem("nonewormwatch", "");
  }
  let noWormWatch = $state(nww !== "");
  const dp = localStorage.getItem("displayPing");
  let displayPing = $state(dp !== null);
  // effect is ok for this because we use checkbox, however prefer onchange
  // for the range because we don't want to commit all the intermediate to
  // localStorage
  $effect(() => {
    localStorage.setItem("new-threads", wantsNewThreads ? "yes" : "");
  });
  $effect(() => {
    localStorage.setItem("nonewormwatch", noWormWatch ? "yes" : "");
  });
  $effect(() => {
    if (displayPing) {
      localStorage.setItem("displayPing", "yes");
    } else {
      localStorage.removeItem("displayPing");
    }
  });
  const fsminval = localStorage.getItem("fs-min");
  let initfsmin = 1;
  if (fsminval !== null) {
    if (fsminval.endsWith("rem")) {
      initfsmin = Number(fsminval.slice(0, fsminval.length - 3));
    }
  }
  let fsmin = $state(initfsmin);
  const fspmaxval = localStorage.getItem("fs-pmax");
  let initfspmax = 1;
  if (fspmaxval !== null) {
    if (fspmaxval.endsWith("rem")) {
      initfspmax = Number(fspmaxval.slice(0, fspmaxval.length - 3));
    }
  }
  let fspmax = $state(initfspmax);
  let local: string = $state("");
  let echo: string = $state("");
</script>

<h2>font size</h2>
<div>
  <div
    class="tx dark"
    id="abc"
    style="font-size:{`${fsmin}rem`};--accent: var(--primary); --accentl: var(--primaryl)"
  >
    <div class="header">
      <span class="nick">wanderer</span>
      <button class="reply"> #abc</button>
      <time class="time" datetime="1970-01-01T00:00:00Z"></time>
    </div>
    <div class="body">
      <div>hello world this is an earlier message</div>
    </div>
    <div class="footer"></div>
  </div>
  <div
    class="tx dark"
    id="def"
    style="font-size:{`${fsmin + fspmax}rem`};--accent: var(--primary); --accentl: var(--primaryl)"
  >
    <div class="header">
      <span class="nick">wanderer</span>
      <button class="reply"> #def</button>
      <time class="time" datetime="1970-01-02T00:00:00Z"></time>
    </div>
    <div class="body">
      <div>hello world this is a later message</div>
    </div>
    <div class="footer"></div>
  </div>
</div>
<div>
  <input
    type="range"
    min="0.5"
    max="1.5"
    step="0.05"
    bind:value={fsmin}
    onchange={() => {
      localStorage.setItem("fs-min", `${fsmin}rem`);
    }}
    id="fs-min"
  />
  <label for="fs-min">min font size in thread (old posts)</label>
</div>
<div>
  <input
    type="range"
    min="0"
    max="1"
    step="0.05"
    bind:value={fspmax}
    onchange={() => {
      localStorage.setItem("fs-pmax", `${fspmax}rem`);
    }}
    id="fs-min"
  />
  <label for="fs-pmax"
    >how much to add to make the max font size in thread (new posts)</label
  >
</div>
<button
  onclick={() => {
    fsmin = 1;
    fspmax = 1;
    localStorage.setItem("fs-min", `${fsmin}rem`);
    localStorage.setItem("fs-pmax", `${fspmax}rem`);
  }}>reset font sizes to recommended!</button
>
<h2>notifications</h2>
<input type="checkbox" bind:checked={wantsNewThreads} id="wants-new-threads" />
<label for="wants-new-threads"
  >notify locally whenever new threads are created? (requires refresh to go into
  effect)</label
>
<h2>volume</h2>
<VolumeSettings />
<h2>thread features</h2>
<input type="checkbox" bind:checked={displayPing} id="display-ping" />
<label for="display-ping"
  >display your ping while in thread if you're interested or maybe for debugging
  (requires refresh)</label
>
<input type="checkbox" bind:checked={noWormWatch} id="no-worm-watch" />
<label for="no-worm-watch">disable worm watch</label>
<h2>debug</h2>
<textarea bind:value={local}></textarea>
<textarea bind:value={echo}></textarea>
<span
  style="font-size:{`${fsmin + fspmax}rem`}; white-space: pre-wrap; --accentl:#ff000080;"
>
  {@html diffzizz(local, echo)}
</span>
