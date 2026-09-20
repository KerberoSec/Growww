/// Secure storage interface for sensitive tokens & keys (Prompt 525 / Prompt 521)
abstract class ISecureStorageService {
  Future<String?> readString({required String key});
  Future<void> writeString({required String key, required String value});
  Future<void> deleteKey({required String key});
  Future<void> clearAll();
}

class InMemorySecureStorageService implements ISecureStorageService {
  final Map<String, String> _storage = {};

  @override
  Future<String?> readString({required String key}) async {
    return _storage[key];
  }

  @override
  Future<void> writeString({required String key, required String value}) async {
    _storage[key] = value;
  }

  @override
  Future<void> deleteKey({required String key}) async {
    _storage.remove(key);
  }

  @override
  Future<void> clearAll() async {
    _storage.clear();
  }
}
