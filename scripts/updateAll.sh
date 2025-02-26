#!/bin/bash
listings=(
  "listing_0043_immediate_movs.txt"
  "listing_0044_register_movs.txt"
  "listing_0045_challenge_register_movs.txt"
  "listing_0046_add_sub_cmp.txt"
  "listing_0047_challenge_flags.txt"
  "listing_0048_ip_register.txt"
  "listing_0049_conditional_jumps.txt"
  "listing_0050_challenge_jumps.txt"
  "listing_0051_memory_mov.txt"
  "listing_0052_memory_add_loop.txt"
  "listing_0053_add_loop_challenge.txt"
  "listing_0054_draw_rectangle.txt"
  "listing_0055_challenge_rectangle.txt"
  "listing_0056_estimating_cycles.txt"
  "listing_0057_challenge_cycles.txt"
)
for i in "${listings[@]}"
do
  if ! "go" run "./scripts/convertHomeWorkOutput.go" "https://raw.githubusercontent.com/cmuratori/computer_enhance/refs/heads/main/perfaware/part1/$i" "./tests/data/simulation/outputs/$i"
  then
    echo "error converting $i"
  fi
done