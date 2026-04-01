import Container from "./Container";

export default function Footer({
  rightLink,
}: {
  rightLink?: { label: string; href: string };
}) {
  return (
    <footer className="border-t border-border py-8">
      <Container>
        <div className="flex flex-wrap items-center justify-between gap-3 max-xs:flex-col max-xs:items-start">
          <p className="text-[13px] text-muted">
            Cookey for humans and agents.
          </p>
          {rightLink && (
            <a
              href={rightLink.href}
              className="text-[13px] text-muted no-underline transition-colors duration-150 hover:text-ink"
            >
              {rightLink.label}
            </a>
          )}
        </div>
      </Container>
    </footer>
  );
}
