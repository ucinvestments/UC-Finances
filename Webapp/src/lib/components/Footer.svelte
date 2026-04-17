<script lang="ts">
  import Icon from "@iconify/svelte";

  interface FooterLink {
    href: string;
    label: string;
    external?: boolean;
  }

  interface CryptoAddress {
    label: string;
    address: string;
  }

  let {
    appName = "UC Data",
    description = "University of California data transparency tool.",
    links = [] as FooterLink[],
    contactEmail = "contact@example.com",
    dataSources = "",
    cryptoAddresses = [] as CryptoAddress[],
    altContactEmail = "",
    copyrightHolder = "",
    affiliationDisclaimer = "",
  }: {
    appName?: string;
    description?: string;
    links?: FooterLink[];
    contactEmail?: string;
    dataSources?: string;
    cryptoAddresses?: CryptoAddress[];
    altContactEmail?: string;
    copyrightHolder?: string;
    affiliationDisclaimer?: string;
  } = $props();

  let donateOpen = $state(false);

  function toggleDonate() {
    donateOpen = !donateOpen;
  }

  const emailDomain = $derived(contactEmail.split("@")[1] ?? "");
  const holder = $derived(copyrightHolder || appName);
  const year = new Date().getFullYear();
</script>

<footer class="footer">
  <div class="footer-content">
    <div class="footer-section">
      <h4 class="footer-title">{appName}</h4>
      <p class="footer-text">{description}</p>
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
        {#if emailDomain}
          <a href="mailto:press@{emailDomain}" class="footer-link">Submit Research</a>
          <a href="mailto:dev@{emailDomain}" class="footer-link">Development</a>
        {/if}
      </div>
    </div>

    {#if cryptoAddresses.length > 0}
      <div class="footer-section">
        <h4 class="footer-title">Support This Project</h4>
        <p class="footer-text">
          Self-funded. Donations help cover hosting and API costs.
        </p>
        <div class="donation-links">
          <button class="donate-button" onclick={toggleDonate}>
            <Icon icon="mdi:heart" class="donate-icon" />
            Donate
          </button>
          {#if donateOpen}
            <div class="donation-info">
              {#each cryptoAddresses as crypto}
                <div class="crypto-address">
                  <strong>{crypto.label}:</strong>
                  <code class="address">{crypto.address}</code>
                </div>
              {/each}
              {#if altContactEmail}
                <p class="alt-contact">
                  Or contact <a href="mailto:{altContactEmail}">us</a> for alternatives.
                </p>
              {/if}
            </div>
          {/if}
        </div>
      </div>
    {/if}

    {#if dataSources}
      <div class="footer-section">
        <h4 class="footer-title">Data Sources</h4>
        <p class="footer-text">{@html dataSources}</p>
      </div>
    {/if}
  </div>

  <div class="footer-bottom">
    <p>
      &copy; {year} {holder}. Educational purposes only.
      {#if affiliationDisclaimer}
        <br />{affiliationDisclaimer}
      {/if}
    </p>
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
