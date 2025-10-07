<script lang="ts">
  import Icon from "@iconify/svelte";
  import { onMount } from "svelte";

  export let appName: string = "UC Data";
  export let description: string = "University of California data transparency tool.";
  export let links: Array<{href: string, label: string, external?: boolean}> = [];
  export let contactEmail: string = "contact@example.com";
  export let lastUpdated: string = "2024";
  export let dataAccuracy: string = "Varies";

  function toggleDonationInfo() {
    const donationInfo = document.getElementById('donationInfo');
    if (donationInfo) {
      donationInfo.style.display = donationInfo.style.display === 'none' ? 'block' : 'none';
    }
  }

  onMount(() => {
    // Make toggle function globally available for the donation button
    window.toggleDonationInfo = toggleDonationInfo;
  });
</script>

<footer class="footer">
  <div class="footer-content">
    <div class="footer-section">
      <h4 class="footer-title">{appName}</h4>
      <p class="footer-text">
        {description}
      </p>
    </div>

    <div class="footer-section">
      <h4 class="footer-title">Quick Links</h4>
      <div class="footer-links">
        {#each links as link}
          <a
            href={link.href}
            class="footer-link"
            target={link.external ? "_blank" : undefined}
            rel={link.external ? "noopener noreferrer" : undefined}
          >
            {link.label}
            {#if link.external}
              <Icon icon="mdi:open-in-new" class="footer-external" />
            {/if}
          </a>
        {/each}
      </div>
    </div>

    <div class="footer-section">
      <h4 class="footer-title">Contact</h4>
      <div class="footer-links">
        <a href="mailto:{contactEmail}" class="footer-link">Contact Us</a>
        <a href="mailto:press@{contactEmail.split('@')[1]}" class="footer-link">Submit Research</a>
        <a href="mailto:dev@{contactEmail.split('@')[1]}" class="footer-link">Development</a>
      </div>
    </div>

    <div class="footer-section">
      <h4 class="footer-title">Support This Project</h4>
      <p class="footer-text">
        This project is self-funded by Stephen. Donations help cover hosting and API costs.
      </p>
      <div class="donation-links">
        <button class="donate-button" on:click={toggleDonationInfo}>
          <Icon icon="mdi:heart" class="donate-icon" />
          Donate
        </button>
        <div class="donation-info" id="donationInfo" style="display: none;">
          <div class="crypto-address">
            <strong>ETH:</strong>
            <code class="address">0x623c7559ddC51BAf15Cc81bf5bc13c0B0EA14c01</code>
          </div>
          <div class="crypto-address">
            <strong>XMR:</strong>
            <code class="address">44bvXALNkxUgSkGChKQPnj79v6JwkeYEkGijgKyp2zRq3EiuL6oewAv5u2c7FN7jbN1z7uj1rrPfL77bbsJ3cC8U2ADFoTj</code>
          </div>
          <p class="alt-contact">
            Or contact <a href="mailto:sdokita@berkeley.edu">Stephen</a> for alternatives.
          </p>
        </div>
      </div>
    </div>

    <div class="footer-section">
      <h4 class="footer-title">Data Sources</h4>
      <p class="footer-text">
        Last updated: {lastUpdated}<br />
        Data accuracy: {dataAccuracy}
      </p>
    </div>
  </div>

  <div class="footer-bottom">
    <p>&copy; 2024 {appName}. Educational purposes only.</p>
  </div>
</footer>

<style>
  .footer {
    background: linear-gradient(180deg, var(--bg-secondary) 0%, white 100%);
    border-top: 1px solid var(--border);
    margin-top: 4rem;
    padding: 3rem 0 1.5rem;
  }

  .footer-content {
    max-width: 1400px;
    margin: 0 auto;
    padding: 0 2rem;
    display: grid;
    grid-template-columns: 2fr 1fr 1fr 1fr 1fr;
    gap: 2rem;
    margin-bottom: 2rem;
  }

  .footer-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .footer-title {
    font-family: "Space Grotesk", sans-serif;
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--pri);
    margin: 0;
  }

  .footer-text {
    color: var(--text-secondary);
    line-height: 1.6;
    font-size: 0.925rem;
    margin: 0;
  }

  .footer-links {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .footer-link {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    color: var(--text-secondary);
    text-decoration: none;
    font-size: 0.925rem;
    transition: all 0.2s ease;
    width: fit-content;
  }

  .footer-link:hover {
    color: var(--founder);
    transform: translateX(4px);
  }

  :global(.footer-external) {
    font-size: 0.75rem;
  }

  .footer-bottom {
    max-width: 1400px;
    margin: 0 auto;
    padding: 2rem 2rem 0;
    border-top: 1px solid var(--border);
    text-align: center;
  }

  .footer-bottom p {
    color: var(--text-secondary);
    font-size: 0.875rem;
    margin: 0;
  }

  /* Donation Styles */
  .donation-links {
    margin-top: 1rem;
  }

  .donate-button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.625rem 1rem;
    background: linear-gradient(135deg, var(--golden-gate), var(--sec));
    color: white;
    border: none;
    border-radius: 0.5rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.3s ease;
    font-size: 0.875rem;
  }

  .donate-button:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
  }

  :global(.donate-icon) {
    font-size: 1rem;
    color: white;
  }

  .donation-info {
    margin-top: 1rem;
    padding: 1rem;
    background: var(--bg-secondary);
    border-radius: 0.5rem;
    border: 1px solid var(--border);
  }

  .crypto-address {
    margin-bottom: 0.75rem;
  }

  .crypto-address strong {
    display: block;
    color: var(--pri);
    font-size: 0.875rem;
    margin-bottom: 0.25rem;
  }

  .address {
    display: block;
    background: white;
    padding: 0.5rem;
    border-radius: 0.25rem;
    font-family: "JetBrains Mono", monospace;
    font-size: 0.75rem;
    word-break: break-all;
    border: 1px solid var(--border);
    color: var(--text-primary);
    cursor: text;
    user-select: all;
  }

  .alt-contact {
    font-size: 0.875rem;
    color: var(--text-secondary);
    margin: 0.75rem 0 0;
  }

  .alt-contact a {
    color: var(--founder);
    text-decoration: none;
    font-weight: 500;
  }

  .alt-contact a:hover {
    color: var(--pri);
    text-decoration: underline;
  }

  @media (max-width: 768px) {
    .footer-content {
      grid-template-columns: 1fr;
      gap: 2rem;
      padding: 0 1.5rem;
    }

    .address {
      font-size: 0.7rem;
    }

    .footer-section {
      text-align: center;
    }

    .footer-links {
      align-items: center;
    }
  }
</style>