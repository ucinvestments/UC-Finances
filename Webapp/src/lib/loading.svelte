<script lang="ts">
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";

  let mounted = false;

  onMount(() => {
    mounted = true;
  });
</script>

{#if mounted}
  <div class="loading-container" in:fade={{ duration: 300 }}>
    <div class="loading-chart">
      <svg viewBox="0 0 120 120" class="chart-svg">
        <!-- Background circle -->
        <circle
          cx="60"
          cy="60"
          r="45"
          fill="none"
          stroke="var(--border)"
          stroke-width="3"
          opacity="0.3"
        />

        <!-- Animated segments -->
        <circle
          cx="60"
          cy="60"
          r="45"
          fill="none"
          stroke="var(--founder)"
          stroke-width="8"
          stroke-linecap="round"
          stroke-dasharray="70 213"
          class="segment segment-1"
          transform="rotate(-90 60 60)"
        />

        <circle
          cx="60"
          cy="60"
          r="45"
          fill="none"
          stroke="var(--sec)"
          stroke-width="8"
          stroke-linecap="round"
          stroke-dasharray="50 233"
          stroke-dashoffset="-70"
          class="segment segment-2"
          transform="rotate(-90 60 60)"
        />

        <circle
          cx="60"
          cy="60"
          r="45"
          fill="none"
          stroke="var(--pri)"
          stroke-width="8"
          stroke-linecap="round"
          stroke-dasharray="40 243"
          stroke-dashoffset="-120"
          class="segment segment-3"
          transform="rotate(-90 60 60)"
        />

        <circle
          cx="60"
          cy="60"
          r="45"
          fill="none"
          stroke="var(--lawrence)"
          stroke-width="8"
          stroke-linecap="round"
          stroke-dasharray="30 253"
          stroke-dashoffset="-160"
          class="segment segment-4"
          transform="rotate(-90 60 60)"
        />

        <!-- Center logo -->
        <g class="center-logo">
          <circle
            cx="60"
            cy="60"
            r="20"
            fill="var(--bg)"
            stroke="var(--border)"
            stroke-width="2"
          />
          <text
            x="60"
            y="66"
            text-anchor="middle"
            class="logo-text"
            font-family="Space Grotesk, sans-serif"
            font-weight="700"
            font-size="14"
            fill="var(--pri)">UC</text
          >
        </g>
      </svg>

      <!-- Pulsing dots -->
      <div class="loading-dots">
        <div class="dot dot-1"></div>
        <div class="dot dot-2"></div>
        <div class="dot dot-3"></div>
      </div>
    </div>

    <div class="loading-text">
      <p class="loading-message">Loading holdings data...</p>
    </div>
  </div>
{/if}

<style>
  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 300px;
    padding: 2rem;
  }

  .loading-chart {
    position: relative;
    margin-bottom: 1.5rem;
  }

  .chart-svg {
    width: 120px;
    height: 120px;
    filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.1));
  }

  .segment {
    animation: rotate 2s linear infinite;
    transform-origin: 60px 60px;
  }

  .segment-1 {
    animation-delay: 0s;
  }

  .segment-2 {
    animation-delay: -0.5s;
  }

  .segment-3 {
    animation-delay: -1s;
  }

  .segment-4 {
    animation-delay: -1.5s;
  }

  .center-logo {
    animation: pulse 2s ease-in-out infinite;
  }

  .logo-text {
    animation: fadeInOut 2s ease-in-out infinite;
  }

  .loading-dots {
    position: absolute;
    bottom: -30px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    gap: 0.5rem;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--founder);
    animation: bounce 1.4s ease-in-out infinite;
  }

  .dot-1 {
    animation-delay: 0s;
  }

  .dot-2 {
    animation-delay: 0.2s;
  }

  .dot-3 {
    animation-delay: 0.4s;
  }

  .loading-text {
    text-align: center;
  }

  .loading-message {
    font-size: 1rem;
    color: var(--text-secondary);
    font-weight: 500;
    margin: 0;
    margin-top: 1rem;
    animation: fadeInOut 2s ease-in-out infinite;
  }

  @keyframes rotate {
    0% {
      transform: rotate(-90deg);
    }
    100% {
      transform: rotate(270deg);
    }
  }

  @keyframes pulse {
    0%,
    100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.05);
      opacity: 0.9;
    }
  }

  @keyframes fadeInOut {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.6;
    }
  }

  @keyframes bounce {
    0%,
    80%,
    100% {
      transform: scale(0.8);
      opacity: 0.5;
    }
    40% {
      transform: scale(1.2);
      opacity: 1;
    }
  }

  @media (max-width: 640px) {
    .chart-svg {
      width: 100px;
      height: 100px;
    }

    .loading-message {
      font-size: 0.925rem;
    }

    .loading-container {
      min-height: 250px;
      padding: 1.5rem;
    }
  }
</style>