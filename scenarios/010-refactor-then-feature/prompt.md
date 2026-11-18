Before we roll out a promo, remove the duplicate item validation in the order service. I want to be able to cherry-pick that cleanup without changing how Quote or Place behaves.

Then let checkout pass an optional discount code to `shop order`. SAVE10 takes 10% off the whole order, no code charges the usual total, and an unrecognized nonempty code should reject the order without reserving anything. Keep the cleanup and the promo in separate local commits; don't open a PR.
