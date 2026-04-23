#include <iostream>
#include <fstream>
#include <sstream>
#include <string>
#include <vector>
#include <map>
#include <filesystem>

struct LineResult {
    bool valid;
    std::string level;
    std::string message;
};

struct FileStats {
    std::string filename;
    int errorCount;
    int warnCount;
    int infoCount;
    int debugCount;
    int parseErrors;
};

LineResult parseLine(const std::string& line) {
    LineResult result;
    result.valid = false;

    if (line.empty()) {
        return result;
    }

    if (line[0] != '[') {
        return result;
    }

    size_t closeBracket = line.find(']');
    if (closeBracket == std::string::npos) {
        return result;
    }

    std::string level = line.substr(1, closeBracket - 1);

    if (level != "ERROR" && level != "WARN" && level != "INFO" && level != "DEBUG") {
        return result;
    }

    if (closeBracket + 1 >= line.size()) {
        return result;
    }

    if (line[closeBracket + 1] != ' ') {
        return result;
    }

    result.valid = true;
    result.level = level;
    result.message = line.substr(closeBracket + 2);
    return result;
}

FileStats aggregateFileStats(const std::string& filepath) {
    FileStats stats;
    stats.filename = std::filesystem::path(filepath).filename().string();
    stats.errorCount = 0;
    stats.warnCount = 0;
    stats.infoCount = 0;
    stats.debugCount = 0;
    stats.parseErrors = 0;

    std::ifstream file(filepath);
    if (!file.is_open()) {
        stats.parseErrors = -1; // Signal that file could not be opened
        return stats;
    }

    std::string line;
    while (std::getline(file, line)) {
        if (line.empty()) {
            continue;
        }

        LineResult result = parseLine(line);
        if (!result.valid) {
            stats.parseErrors++;
        } else {
            if (result.level == "ERROR") {
                stats.errorCount++;
            } else if (result.level == "WARN") {
                stats.warnCount++;
            } else if (result.level == "INFO") {
                stats.infoCount++;
            } else if (result.level == "DEBUG") {
                stats.debugCount++;
            }
        }
    }

    file.close();
    return stats;
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        std::cerr << "Usage: " << argv[0] << " <logfile1> [logfile2] ..." << std::endl;
        return 1;
    }

    std::vector<FileStats> allStats;
    bool anyError = false;

    for (int i = 1; i < argc; i++) {
        std::string filepath = argv[i];
        FileStats stats = aggregateFileStats(filepath);

        if (stats.parseErrors == -1) {
            std::cerr << "Error: Cannot read file: " << filepath << std::endl;
            anyError = true;
            continue;
        }

        allStats.push_back(stats);

        std::cout << stats.filename << ": "
                  << "ERROR=" << stats.errorCount << " "
                  << "WARN=" << stats.warnCount << " "
                  << "INFO=" << stats.infoCount << " "
                  << "DEBUG=" << stats.debugCount << " "
                  << "parse_errors=" << stats.parseErrors << std::endl;
    }

    int totalFiles = (int)allStats.size();
    int totalLines = 0;
    int totalParseErrors = 0;

    for (const auto& stats : allStats) {
        totalLines += stats.errorCount + stats.warnCount + stats.infoCount + stats.debugCount + stats.parseErrors;
        totalParseErrors += stats.parseErrors;
    }

    std::cout << "TOTAL: "
              << "files=" << totalFiles << " "
              << "lines=" << totalLines << " "
              << "parse_errors=" << totalParseErrors << std::endl;

    return anyError ? 1 : 0;
}
