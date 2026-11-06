import { z } from "zod";

// Zod checks whether it can compile parsers with eval. The CSP blocks eval, so
// that check would report a violation on every page that validates a form.
// Without compiling, Zod parses the same way, and forms this size are no slower.
z.config({ jitless: true });
