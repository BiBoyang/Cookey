import { useCallback, useState } from "react";
import Nav from "../components/Nav";
import Footer from "../components/Footer";
import Container from "../components/Container";
import Badge from "../components/Badge";
import { Button } from "../components/Button";
import { AGENT_MARKDOWN } from "../data/agentMarkdown";

export default function GetStartedPage() {
  const [copyState, setCopyState] = useState<"idle" | "copied" | "failed">(
    "idle",
  );

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(AGENT_MARKDOWN + "\n");
      setCopyState("copied");
    } catch {
      setCopyState("failed");
    }
    setTimeout(() => setCopyState("idle"), 1800);
  }, []);

  return (
    <div className="bg-bg text-ink font-sans leading-[1.6]">
      <Nav />

      <main>
        <Container>
          <section className="pt-20 pb-16">
            <div className="mb-7">
              <Badge>Agent Handoff</Badge>
            </div>
            <h1 className="mb-[18px] font-bold tracking-[-0.03em] leading-[1.1] text-[clamp(2.2rem,6vw,3.2rem)]">
              Paste this, let your agent handle it.
            </h1>
            <p className="mb-9 max-w-[520px] text-[1.05rem] text-muted">
              This is not an install wizard. Copy the instructions below, paste
              it into your terminal agent, and let it install Cookey.
            </p>

            <Button
              variant="primary"
              onClick={handleCopy}
              data-state={copyState === "copied" ? "copied" : ""}
            >
              {copyState === "copied"
                ? "Copied"
                : copyState === "failed"
                  ? "Copy failed"
                  : "Copy for Agents"}
            </Button>

            <div className="mt-8 overflow-hidden rounded-xl border border-border">
              <div className="overflow-x-auto bg-terminal-bg p-[24px_20px]">
                <pre className="m-0 whitespace-pre-wrap break-words font-mono text-[13px] leading-[1.8] text-muted">
                  {AGENT_MARKDOWN}
                </pre>
              </div>
            </div>
          </section>
        </Container>
      </main>

      <Footer rightLink={{ label: "llms.txt", href: "/llms.txt" }} />
    </div>
  );
}
