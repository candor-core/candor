#include <iostream>
#include <fstream>
#include <sstream>
#include <string>
#include <vector>
#include <cstring>

struct LineResult {
    bool valid;
    std::string level;
};

struct FileStats {
    int error_count;
    int warn_count;
    int info_count;
    int debug_count;
    int parse_errors;
    int total_lines;
};

LineResult parseLine(const std::string& line) {
    LineResult result;
    result.valid = false;
    result.level = "";

    if (line.empty()) {
        return result;
    }

    if (line[0] != '[') {
        return result;
    }

    size_t close_bracket = line.find(']');
    if (close_bracket == std::string::npos) {
        return result;
    }

    std::string level = line.substr(1, close_bracket - 1);

    if (level != "ERROR" && level != "WARN" && level != "INFO" && level != "DEBUG") {
        return result;
    }

    if (close_bracket + 1 >= line.size() || line[close_bracket + 1] != ' ') {
        return result;
    }

    result.valid = true;
    result.level = level;
    return result;
}

FileStats aggregateFileStats(const std::vector<std::string>& lines) {
    FileStats stats;
    stats.error_count = 0;
    stats.warn_count = 0;
    stats.info_count = 0;
    stats.debug_count = 0;
    stats.parse_errors = 0;
    stats.total_lines = 0;

    for (size_t i = 0; i < lines.size(); i++) {
        const std::string& line = lines[i];
        if (line.empty()) {
            continue;
        }

        stats.total_lines++;
        LineResult result = parseLine(line);

        if (!result.valid) {
            stats.parse_errors++;
        } else if (result.level == "ERROR") {
            stats.error_count++;
        } else if (result.level == "WARN") {
            stats.warn_count++;
        } else if (result.level == "INFO") {
            stats.info_count++;
        } else if (result.level == "DEBUG") {
            stats.debug_count++;
        }
    }

    return stats;
}

static std::string getFilename(const std::string& filepath) {
    size_t pos = filepath.find_last_of("/\\");
    if (pos == std::string::npos) {
        return filepath;
    }
    return filepath.substr(pos + 1);
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        std::cerr << "Usage: " << argv[0] << " <logfile1> [logfile2 ...]" << std::endl;
        return 1;
    }

    int total_files = 0;
    int total_lines = 0;
    int total_parse_errors = 0;

    for (int i = 1; i < argc; i++) {
        std::string filepath = argv[i];
        std::ifstream file(filepath.c_str());

        if (!file.is_open()) {
            std::cerr << "Error: Cannot open file: " << filepath << std::endl;
            return 1;
        }

        std::vector<std::string> lines;
        std::string line;
        while (std::getline(file, line)) {
            lines.push_back(line);
        }

        file.close();

        FileStats stats = aggregateFileStats(lines);

        std::string filename = getFilename(filepath);

        std::cout << filename << ": "
                  << "ERROR=" << stats.error_count << " "
                  << "WARN=" << stats.warn_count << " "
                  << "INFO=" << stats.info_count << " "
                  << "DEBUG=" << stats.debug_count << " "
                  << "parse_errors=" << stats.parse_errors << std::endl;

        total_files++;
        total_lines += stats.total_lines;
        total_parse_errors += stats.parse_errors;
    }

    std::cout << "TOTAL: files=" << total_files
              << " lines=" << total_lines
              << " parse_errors=" << total_parse_errors << std::endl;

    return 0;
}
