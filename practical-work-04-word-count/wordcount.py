#!/usr/bin/env python3
"""
Word Count using MapReduce
Practical Work 4
"""

import os
import sys
import re
from collections import defaultdict
from multiprocessing import Pool
import argparse

def map_function(filename):
    """Mapper: emit (word, 1) for each word"""
    word_counts = defaultdict(int)
    
    try:
        with open(filename, 'r', encoding='utf-8') as f:
            for line in f:
                # Tokenize by whitespace and punctuation
                words = re.findall(r'\b\w+\b', line.lower())
                for word in words:
                    word_counts[word] += 1
    except Exception as e:
        print(f"Error reading {filename}: {e}")
    
    return word_counts

def reduce_function(word_counts_list):
    """Reducer: sum counts for each word"""
    final_counts = defaultdict(int)
    
    for word_counts in word_counts_list:
        for word, count in word_counts.items():
            final_counts[word] += count
    
    return final_counts

def mapreduce(input_dir, output_dir, num_mappers=2):
    """Run MapReduce word count"""
    # Get input files
    input_files = []
    for filename in os.listdir(input_dir):
        filepath = os.path.join(input_dir, filename)
        if os.path.isfile(filepath):
            input_files.append(filepath)
    
    if not input_files:
        print(f"No files found in {input_dir}")
        return
    
    print(f"Found {len(input_files)} input files")
    print(f"Using {num_mappers} mappers")
    
    # Map phase
    print("\nMap phase...")
    with Pool(num_mappers) as pool:
        map_results = pool.map(map_function, input_files)
    
    # Reduce phase
    print("Reduce phase...")
    final_counts = reduce_function(map_results)
    
    # Write output
    os.makedirs(output_dir, exist_ok=True)
    output_file = os.path.join(output_dir, "part-00000")
    
    with open(output_file, 'w', encoding='utf-8') as f:
        for word, count in sorted(final_counts.items()):
            f.write(f"{word}\t{count}\n")
    
    print(f"\nWord count completed!")
    print(f"Output written to: {output_file}")
    print(f"Total unique words: {len(final_counts)}")
    
    # Show top 10 words
    print("\nTop 10 words:")
    sorted_words = sorted(final_counts.items(), key=lambda x: x[1], reverse=True)
    for word, count in sorted_words[:10]:
        print(f"  {word}: {count}")

def main():
    parser = argparse.ArgumentParser(description='Word Count MapReduce')
    parser.add_argument('input_dir', help='Input directory')
    parser.add_argument('output_dir', help='Output directory')
    parser.add_argument('--mappers', type=int, default=2, help='Number of mappers')
    args = parser.parse_args()
    
    mapreduce(args.input_dir, args.output_dir, args.mappers)

if __name__ == '__main__':
    main()

