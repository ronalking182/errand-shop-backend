// Mobile App - Custom Requests Screen Example
// This demonstrates how the mobile app should integrate with the dashboard

import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  Alert,
  RefreshControl,
  Modal,
  ScrollView
} from 'react-native';
import PushNotification from 'react-native-push-notification';

// API Service for mobile app
class MobileApiService {
  constructor() {
    this.baseURL = 'http://localhost:9090/api/v1';
  }

  async request(endpoint, options = {}) {
    const token = await this.getAuthToken(); // Get user token from secure storage
    
    const headers = {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers
    };

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers
    });

    if (response.status === 401) {
      // Handle unauthorized - redirect to login
      this.handleUnauthorized();
      throw new Error('Unauthorized');
    }

    return response;
  }

  async getAuthToken() {
    // In real app, get from secure storage (AsyncStorage, Keychain, etc.)
    return global.userToken;
  }

  handleUnauthorized() {
    // Navigate to login screen
    // NavigationService.navigate('Login');
  }

  // Custom Requests API methods
  async getMyCustomRequests() {
    const response = await this.request('/custom-requests');
    return response.json();
  }

  async createCustomRequest(requestData) {
    const response = await this.request('/custom-requests', {
      method: 'POST',
      body: JSON.stringify(requestData)
    });
    return response.json();
  }

  async acceptQuote(requestId, quoteId) {
    const response = await this.request(`/custom-requests/${requestId}/accept-quote`, {
      method: 'POST',
      body: JSON.stringify({ quoteId })
    });
    return response.json();
  }

  async declineQuote(requestId, quoteId, reason) {
    const response = await this.request(`/custom-requests/${requestId}/decline-quote`, {
      method: 'POST',
      body: JSON.stringify({ quoteId, reason })
    });
    return response.json();
  }

  async sendMessage(requestId, message) {
    const response = await this.request(`/custom-requests/${requestId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content: message, senderType: 'customer' })
    });
    return response.json();
  }
}

const apiService = new MobileApiService();

// Custom Request Card Component
const CustomRequestCard = ({ request, onPress, onAcceptQuote, onDeclineQuote }) => {
  const getStatusColor = (status) => {
    const colors = {
      'submitted': '#007bff',
      'under_review': '#ffc107',
      'quote_sent': '#17a2b8',
      'customer_accepted': '#28a745',
      'approved': '#28a745',
      'cancelled': '#dc3545'
    };
    return colors[status] || '#6c757d';
  };

  const formatStatus = (status) => {
    return status.replace('_', ' ').toUpperCase();
  };

  return (
    <TouchableOpacity style={styles.requestCard} onPress={onPress}>
      <View style={styles.cardHeader}>
        <Text style={styles.requestId}>#{request.id.substring(0, 8)}</Text>
        <View style={[styles.statusBadge, { backgroundColor: getStatusColor(request.status) }]}>
          <Text style={styles.statusText}>{formatStatus(request.status)}</Text>
        </View>
      </View>
      
      <Text style={styles.itemCount}>
        {request.items?.length || 0} item(s) requested
      </Text>
      
      <Text style={styles.createdDate}>
        Created: {new Date(request.createdAt).toLocaleDateString()}
      </Text>
      
      {request.activeQuote && (
        <View style={styles.quoteSection}>
          <Text style={styles.quoteLabel}>Quoted Price:</Text>
          <Text style={styles.quoteAmount}>
            ₦{(request.activeQuote.grandTotal / 100).toFixed(2)}
          </Text>
          
          {request.status === 'quote_sent' && (
            <View style={styles.quoteActions}>
              <TouchableOpacity 
                style={[styles.button, styles.acceptButton]}
                onPress={() => onAcceptQuote(request.id, request.activeQuote.id)}
              >
                <Text style={styles.buttonText}>Accept Quote</Text>
              </TouchableOpacity>
              
              <TouchableOpacity 
                style={[styles.button, styles.declineButton]}
                onPress={() => onDeclineQuote(request.id, request.activeQuote.id)}
              >
                <Text style={styles.buttonText}>Decline</Text>
              </TouchableOpacity>
            </View>
          )}
        </View>
      )}
    </TouchableOpacity>
  );
};

// Quote Details Modal
const QuoteDetailsModal = ({ visible, quote, onClose, onAccept, onDecline }) => {
  if (!quote) return null;

  return (
    <Modal visible={visible} animationType="slide" presentationStyle="pageSheet">
      <View style={styles.modalContainer}>
        <View style={styles.modalHeader}>
          <Text style={styles.modalTitle}>Quote Details</Text>
          <TouchableOpacity onPress={onClose}>
            <Text style={styles.closeButton}>✕</Text>
          </TouchableOpacity>
        </View>
        
        <ScrollView style={styles.modalContent}>
          <Text style={styles.sectionTitle}>Items</Text>
          {quote.items?.map((item, index) => (
            <View key={index} style={styles.quoteItem}>
              <Text style={styles.itemName}>{item.name}</Text>
              <Text style={styles.itemPrice}>₦{(item.quotedPrice / 100).toFixed(2)}</Text>
              {item.adminNotes && (
                <Text style={styles.itemNotes}>Note: {item.adminNotes}</Text>
              )}
            </View>
          ))}
          
          <Text style={styles.sectionTitle}>Fees</Text>
          <View style={styles.feesContainer}>
            <View style={styles.feeRow}>
              <Text>Delivery Fee:</Text>
              <Text>₦{(quote.fees?.delivery / 100 || 0).toFixed(2)}</Text>
            </View>
            <View style={styles.feeRow}>
              <Text>Service Fee:</Text>
              <Text>₦{(quote.fees?.service / 100 || 0).toFixed(2)}</Text>
            </View>
            <View style={styles.feeRow}>
              <Text>Packaging Fee:</Text>
              <Text>₦{(quote.fees?.packaging / 100 || 0).toFixed(2)}</Text>
            </View>
          </View>
          
          <View style={styles.totalContainer}>
            <Text style={styles.totalText}>
              Total: ₦{(quote.grandTotal / 100).toFixed(2)}
            </Text>
          </View>
          
          {quote.validUntil && (
            <Text style={styles.validUntil}>
              Valid until: {new Date(quote.validUntil).toLocaleString()}
            </Text>
          )}
        </ScrollView>
        
        <View style={styles.modalActions}>
          <TouchableOpacity 
            style={[styles.button, styles.declineButton]}
            onPress={onDecline}
          >
            <Text style={styles.buttonText}>Decline</Text>
          </TouchableOpacity>
          
          <TouchableOpacity 
            style={[styles.button, styles.acceptButton]}
            onPress={onAccept}
          >
            <Text style={styles.buttonText}>Accept Quote</Text>
          </TouchableOpacity>
        </View>
      </View>
    </Modal>
  );
};

// Main Custom Requests Screen
const CustomRequestsScreen = ({ navigation }) => {
  const [requests, setRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [selectedQuote, setSelectedQuote] = useState(null);
  const [quoteModalVisible, setQuoteModalVisible] = useState(false);

  useEffect(() => {
    loadRequests();
    setupPushNotifications();
  }, []);

  const setupPushNotifications = () => {
    PushNotification.configure({
      onNotification: function(notification) {
        console.log('Notification received:', notification);
        
        if (notification.data?.type === 'quote_sent') {
          // Refresh requests to show new quote
          loadRequests();
          
          // Show alert about new quote
          Alert.alert(
            'New Quote Available!',
            'You have received a new quote for your custom request.',
            [
              { text: 'View Later', style: 'cancel' },
              { text: 'View Now', onPress: () => {
                // Navigate to the specific request
                const requestId = notification.data.requestId;
                if (requestId) {
                  viewRequestDetails(requestId);
                }
              }}
            ]
          );
        }
      },
      
      requestPermissions: Platform.OS === 'ios'
    });
  };

  const loadRequests = async () => {
    try {
      setLoading(true);
      const response = await apiService.getMyCustomRequests();
      
      if (response.success) {
        setRequests(response.data.requests || []);
      } else {
        throw new Error(response.message || 'Failed to load requests');
      }
    } catch (error) {
      console.error('Error loading requests:', error);
      Alert.alert('Error', 'Failed to load custom requests. Please try again.');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const onRefresh = () => {
    setRefreshing(true);
    loadRequests();
  };

  const viewRequestDetails = (requestId) => {
    const request = requests.find(r => r.id === requestId);
    if (request && request.activeQuote) {
      setSelectedQuote(request.activeQuote);
      setQuoteModalVisible(true);
    } else {
      // Navigate to request details screen
      navigation.navigate('RequestDetails', { requestId });
    }
  };

  const acceptQuote = async (requestId, quoteId) => {
    try {
      const response = await apiService.acceptQuote(requestId, quoteId);
      
      if (response.success) {
        Alert.alert(
          'Quote Accepted!',
          'Your quote has been accepted. Your order is now being processed.',
          [{ text: 'OK', onPress: () => {
            setQuoteModalVisible(false);
            loadRequests(); // Refresh to show updated status
          }}]
        );
      } else {
        throw new Error(response.message || 'Failed to accept quote');
      }
    } catch (error) {
      console.error('Error accepting quote:', error);
      Alert.alert('Error', 'Failed to accept quote. Please try again.');
    }
  };

  const declineQuote = async (requestId, quoteId) => {
    Alert.prompt(
      'Decline Quote',
      'Please provide a reason for declining this quote:',
      [
        { text: 'Cancel', style: 'cancel' },
        { text: 'Decline', onPress: async (reason) => {
          try {
            const response = await apiService.declineQuote(requestId, quoteId, reason);
            
            if (response.success) {
              Alert.alert(
                'Quote Declined',
                'The quote has been declined. The admin will be notified.',
                [{ text: 'OK', onPress: () => {
                  setQuoteModalVisible(false);
                  loadRequests();
                }}]
              );
            } else {
              throw new Error(response.message || 'Failed to decline quote');
            }
          } catch (error) {
            console.error('Error declining quote:', error);
            Alert.alert('Error', 'Failed to decline quote. Please try again.');
          }
        }}
      ],
      'plain-text'
    );
  };

  const renderRequest = ({ item }) => (
    <CustomRequestCard
      request={item}
      onPress={() => viewRequestDetails(item.id)}
      onAcceptQuote={acceptQuote}
      onDeclineQuote={declineQuote}
    />
  );

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>My Custom Requests</Text>
        <TouchableOpacity 
          style={styles.addButton}
          onPress={() => navigation.navigate('CreateCustomRequest')}
        >
          <Text style={styles.addButtonText}>+ New Request</Text>
        </TouchableOpacity>
      </View>
      
      <FlatList
        data={requests}
        renderItem={renderRequest}
        keyExtractor={(item) => item.id}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        contentContainerStyle={styles.listContainer}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyText}>
              {loading ? 'Loading...' : 'No custom requests yet'}
            </Text>
          </View>
        }
      />
      
      <QuoteDetailsModal
        visible={quoteModalVisible}
        quote={selectedQuote}
        onClose={() => setQuoteModalVisible(false)}
        onAccept={() => {
          if (selectedQuote) {
            const request = requests.find(r => r.activeQuote?.id === selectedQuote.id);
            if (request) {
              acceptQuote(request.id, selectedQuote.id);
            }
          }
        }}
        onDecline={() => {
          if (selectedQuote) {
            const request = requests.find(r => r.activeQuote?.id === selectedQuote.id);
            if (request) {
              declineQuote(request.id, selectedQuote.id);
            }
          }
        }}
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8f9fa'
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    backgroundColor: 'white',
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0'
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#333'
  },
  addButton: {
    backgroundColor: '#007bff',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8
  },
  addButtonText: {
    color: 'white',
    fontWeight: '600'
  },
  listContainer: {
    padding: 16
  },
  requestCard: {
    backgroundColor: 'white',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8
  },
  requestId: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#333'
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12
  },
  statusText: {
    color: 'white',
    fontSize: 12,
    fontWeight: '600'
  },
  itemCount: {
    fontSize: 14,
    color: '#666',
    marginBottom: 4
  },
  createdDate: {
    fontSize: 12,
    color: '#999'
  },
  quoteSection: {
    marginTop: 12,
    padding: 12,
    backgroundColor: '#f8f9fa',
    borderRadius: 8
  },
  quoteLabel: {
    fontSize: 14,
    color: '#666',
    marginBottom: 4
  },
  quoteAmount: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#28a745',
    marginBottom: 12
  },
  quoteActions: {
    flexDirection: 'row',
    gap: 8
  },
  button: {
    flex: 1,
    paddingVertical: 8,
    borderRadius: 6,
    alignItems: 'center'
  },
  acceptButton: {
    backgroundColor: '#28a745'
  },
  declineButton: {
    backgroundColor: '#dc3545'
  },
  buttonText: {
    color: 'white',
    fontWeight: '600'
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingTop: 50
  },
  emptyText: {
    fontSize: 16,
    color: '#666'
  },
  // Modal styles
  modalContainer: {
    flex: 1,
    backgroundColor: 'white'
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0'
  },
  modalTitle: {
    fontSize: 18,
    fontWeight: 'bold'
  },
  closeButton: {
    fontSize: 24,
    color: '#666'
  },
  modalContent: {
    flex: 1,
    padding: 16
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: 'bold',
    marginTop: 16,
    marginBottom: 8
  },
  quoteItem: {
    padding: 12,
    backgroundColor: '#f8f9fa',
    borderRadius: 8,
    marginBottom: 8
  },
  itemName: {
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 4
  },
  itemPrice: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#28a745'
  },
  itemNotes: {
    fontSize: 12,
    color: '#666',
    fontStyle: 'italic',
    marginTop: 4
  },
  feesContainer: {
    backgroundColor: '#f8f9fa',
    padding: 12,
    borderRadius: 8
  },
  feeRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 4
  },
  totalContainer: {
    backgroundColor: '#e8f5e8',
    padding: 16,
    borderRadius: 8,
    marginTop: 16,
    alignItems: 'center'
  },
  totalText: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#28a745'
  },
  validUntil: {
    fontSize: 12,
    color: '#666',
    textAlign: 'center',
    marginTop: 8
  },
  modalActions: {
    flexDirection: 'row',
    padding: 16,
    gap: 12,
    borderTopWidth: 1,
    borderTopColor: '#e0e0e0'
  }
});

export default CustomRequestsScreen;