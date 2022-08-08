# Changelog

## Unreleased

 - **Breaking:** This version contains breaking changes in the AttackKit API.
   - Attack description `category` is now optional and no longer an enumeration.
   - Attack's `targetType` is now optional.
   - `target` is now optional within prepare request bodies.
   - new required configuration value `kind` in preparation for action, checks load test support.
   - Several types were renamed to make it easier to under when to use them.
   - `attack_kit_api.Ptr` was removed from this module to avoid requirement for Go `>1.18`.
   - Enum names for HTTP methods were changed to avoid context specific prefixes.
 - Added the ability to upload artifacts to Steadybit as part of the `prepare`,
   `start`, `status` and `stop` endpoints.

## 0.5.0

 - Support `key_value` attack parameters.

## 0.4.0

 - Support `file` attack parameters.

## 0.3.0

 - Support `options` for attack parameters.

## 0.2.0

 - Support `string_array` as an alias for the `string[]` parameter type.

## 0.1.0

 - Initial release