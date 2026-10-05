# V55 preflight repairs

All changes are isolated research changes. No main timing output collected yet.

- Mechanically cloned model/reference fixture initially failed compilation:
  `cannot use gold (variable of struct type researchpaired.Option) as
  researchpairedfast.Option value in argument to closeOptionPairedV55`.
  CONFIRMED named-type mismatch, not an inference defect. Alias the original
  immutable Config and Option value types; retain independent candidate methods.
  No forecast implementation is delegated to the original model.
- Initial six cloned model tests and the two new cache-fence tests passed before
  the fixture compile attempt. They do not by themselves prove full trajectory
  equivalence or a computational rescue.

Preflight inputs are developmental. V54's diagnostic populations are already
consumed and may test computational equivalence only, not fresh confirmation.
Original V54 outputs, source files and checkpoints remain immutable.

- First prospective run at research/paired-v55-diagnostic terminated1 at vet,
  before allocation or experiment dispatch. Eight model roots and three fixture
  roots passed under race first. Go vet rejected the six inherited unkeyed
  Config literals because the alias is an external package's struct. CONFIRMED
  test syntax issue; rewrite those exact six literals with named fields. No
  parameter changes. Preserve failed source copies/logs/failure.json and use a
  NEW exclusive research/paired-v55-rescue root/freeze for the full run.
