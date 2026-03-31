#!/bin/bash
cpu_info() {
    output_file="cpu.stats"
    tmp_data_file=~/.xterminal/cpu.tmp
    {
      echo "cpu:"
      if [[ -f $tmp_data_file ]]; then
          prev_data=$(cat $tmp_data_file)
          echo "    old:"
          while read -r name user system idle total _; do
              echo "        - name: $name
          loadUser: $user
          loadSystem: $system
          loadIdle: $idle
          loadTotal: $total"
          done <<< "$prev_data"
      fi
      echo "    new:"
      core_num=$(grep -c '^cpu[0-9]' /proc/stat)
      lines_count=$((core_num + 1))
      cpu_lines=$(head -n $lines_count /proc/stat)
      count=0
      while read -r name user nice system idle iowait irq softirq steal guest guest_nice _; do
          user_total=$((user + nice))
          system_total=$((system + irq + softirq + ${steal:-0}))
          idle_total=$((idle + iowait))
          total=$((user_total + system_total + idle_total))
          echo "        - name: $name
          loadUser: $user_total
          loadSystem: $system_total
          loadIdle: $idle_total
          loadTotal: $total"
          if [[ $count -eq 0 ]]; then
              echo "$name $user_total $system_total $idle_total $total" > $tmp_data_file
          else
              echo "$name $user_total $system_total $idle_total $total" >> $tmp_data_file
          fi
          count=$((count + 1))
      done <<< "$cpu_lines"
    } > "$output_file"
}
__main() {
    cpu_info
}
__main
